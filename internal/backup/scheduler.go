// Package backup also implements the auto-backup scheduler.
// Runs daily backups with rotation, all local. Cloud/P2P backup
// is explicitly Phase 12 (Reach), not here.
//
// The scheduler is a small goroutine + ticker. It is started by
// the daemon at boot and stopped on shutdown via the context.
package backup

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Default scheduler settings.
const (
	defaultSchedulerKeepN = 7
)

// SchedulerConfig configures the auto-backup scheduler.
type SchedulerConfig struct {
	// Interval is the time between auto-backups. Default: 24h.
	Interval time.Duration
	// KeepN is the number of recent backups to retain. Older ones
	// are deleted. Default: 7.
	KeepN int
	// RetentionDays is the age-based retention policy: backups older
	// than this many days are pruned. 0 = no age-prune (forever). O3:
	// previously this user config knob was ignored; Rotate now honors it
	// alongside KeepN (a backup is pruned if past KeepN OR older than
	// RetentionDays). Default: 0 here (the daemon passes the user's
	// config value, default 30 in config.Default).
	RetentionDays int
	// BackupDir is the directory where backups are written. If
	// empty, defaults to <data-dir>/backups.
	BackupDir string
	// FirstRunAt is when the scheduler should first run. If zero,
	// it runs immediately at startup (so the user has at least
	// one backup after the first day).
	FirstRunAt time.Time
	// Now is the source of "now" for tests.
	Now func() time.Time
}

// DefaultSchedulerConfig returns the safe defaults.
func DefaultSchedulerConfig() SchedulerConfig {
	return SchedulerConfig{
		Interval:      24 * time.Hour,
		KeepN:         defaultSchedulerKeepN,
		RetentionDays: 0,
		BackupDir:     "",
		FirstRunAt:    time.Time{},
		Now:           time.Now,
	}
}

// Scheduler runs auto-backups on a cadence.
type Scheduler struct {
	cfg  SchedulerConfig
	bm   *Manager
	log  *slog.Logger
	stop chan struct{}
}

// NewScheduler creates a Scheduler.
func NewScheduler(cfg SchedulerConfig, bm *Manager, log *slog.Logger) *Scheduler {
	if cfg.Interval <= 0 {
		cfg.Interval = 24 * time.Hour
	}
	if cfg.KeepN <= 0 {
		cfg.KeepN = defaultSchedulerKeepN
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	// Apply the documented default: empty BackupDir means
	// <data-dir>/backups.
	if cfg.BackupDir == "" && bm != nil && bm.opts.DataDir != "" {
		cfg.BackupDir = ResolveBackupDir(bm.opts.DataDir)
	}
	return &Scheduler{cfg: cfg, bm: bm, log: log, stop: make(chan struct{})}
}

// Run blocks until ctx is canceled. It is safe to call from a
// goroutine. Run performs an initial backup at FirstRunAt (or
// immediately if zero), then loops on cfg.Interval.
func (s *Scheduler) Run(ctx context.Context) {
	if s.log != nil {
		s.log.Info("backup scheduler started",
			"interval", s.cfg.Interval,
			"keep_n", s.cfg.KeepN,
			"backup_dir", s.cfg.BackupDir)
	}
	now := s.cfg.Now()
	if !s.cfg.FirstRunAt.IsZero() {
		// Wait until FirstRunAt, but check ctx and stop frequently.
		wait := s.cfg.FirstRunAt.Sub(now)
		if wait < 0 {
			wait = 0
		}
		select {
		case <-ctx.Done():
			return
		case <-s.stop:
			return
		case <-time.After(wait):
		}
	}

	// Run the first backup immediately, then loop.
	s.tryBackup(ctx)

	t := time.NewTicker(s.cfg.Interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stop:
			return
		case <-t.C:
			s.tryBackup(ctx)
		}
	}
}

// Stop signals the scheduler to exit. Run() returns at the next
// tick or stop.
func (s *Scheduler) Stop() {
	select {
	case <-s.stop:
	default:
		close(s.stop)
	}
}

// tryBackup runs a single backup cycle: create a new archive,
// then rotate. Errors are logged, not returned, because the
// scheduler is best-effort and a backup failure must not bring
// down the daemon.
func (s *Scheduler) tryBackup(ctx context.Context) {
	if err := os.MkdirAll(s.cfg.BackupDir, 0o700); err != nil {
		if s.log != nil {
			s.log.Warn("auto-backup failed", "err", err, "dir", s.cfg.BackupDir)
		}
		return
	}
	now := s.cfg.Now()
	out := filepath.Join(s.cfg.BackupDir, "condura-backup-"+now.Format("2006-01-02T15-04-05Z")+".zip")
	opts := Options{
		DataDir:       s.bm.opts.DataDir,
		MasterKey:     s.bm.opts.MasterKey,
		SchemaVersion: s.bm.opts.SchemaVersion,
		Now:           now,
		Out:           out,
	}
	// Reuse the same Manager so the data dir / master key are
	// the same; the scheduler is just a periodic caller.
	bm := &Manager{opts: opts}
	path, err := bm.Create(ctx)
	if err != nil {
		if s.log != nil {
			s.log.Warn("auto-backup failed", "err", err, "out", out)
		}
		return
	}
	if s.log != nil {
		s.log.Info("auto-backup created", "path", path)
	}
	if err := s.Rotate(); err != nil {
		if s.log != nil {
			s.log.Warn("backup rotation failed", "err", err)
		}
	}
}

// Rotate keeps only the KeepN most recent backups in BackupDir.
// Older ones are deleted. Safe to call on an empty dir.
func (s *Scheduler) Rotate() error {
	dir := s.cfg.BackupDir
	if dir == "" {
		return nil
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	type entry struct {
		name    string
		modTime time.Time
	}
	var backups []entry
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if !strings.HasPrefix(f.Name(), "condura-backup-") || !strings.HasSuffix(f.Name(), ".zip") {
			continue
		}
		info, err := f.Info()
		if err != nil {
			continue
		}
		backups = append(backups, entry{name: info.Name(), modTime: info.ModTime()})
	}
	// Sort newest-first.
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].modTime.After(backups[j].modTime)
	})
	// O3: prune by count (KeepN) AND age (RetentionDays). A backup is
	// pruned if it is past KeepN OR older than RetentionDays. RetentionDays
	// == 0 means no age-prune (forever). Previously RetentionDays was a
	// shipped config knob that was read nowhere.
	now := s.cfg.Now()
	var cutoff time.Time
	if s.cfg.RetentionDays > 0 {
		cutoff = now.Add(-time.Duration(s.cfg.RetentionDays) * 24 * time.Hour)
	}
	for i, b := range backups {
		pruneByCount := i >= s.cfg.KeepN
		pruneByAge := s.cfg.RetentionDays > 0 && b.modTime.Before(cutoff)
		if pruneByCount || pruneByAge {
			if err := os.Remove(filepath.Join(dir, b.name)); err != nil {
				return err
			}
		}
	}
	return nil
}

// Cfg returns the scheduler's resolved configuration (with
// defaults applied). Callers can read this to log the actual
// cadence and backup dir after NewScheduler has filled in
// any missing fields.
func (s *Scheduler) Cfg() SchedulerConfig { return s.cfg }
