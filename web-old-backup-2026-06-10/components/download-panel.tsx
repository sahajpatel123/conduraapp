import { ArrowDownToLine, Info } from "lucide-react";
import { platforms, releaseStates } from "@/lib/site-data";
import { Button } from "@/components/button";
import { Reveal } from "@/components/reveal";

export function PlatformDownloadGrid() {
  return (
    <div className="release-stack">
      {platforms.map((platform, index) => {
        const Icon = platform.icon;
        const statusId = `${platform.name.toLowerCase()}-download-status`;
        return (
          <Reveal key={platform.name} delay={index * 0.06}>
            <article className="release-artifact">
              <div className="download-card-head">
                <div className="download-platform">
                  <span className="platform-icon">
                    <Icon aria-hidden="true" size={22} />
                  </span>
                  <div>
                    <h3>{platform.name}</h3>
                    <p>{platform.version}</p>
                  </div>
                </div>
                <span className="release-pill">
                  {platform.status}
                </span>
              </div>

              <p id={statusId} className="download-detail">
                {platform.detail}
              </p>

              <dl className="download-meta">
                <div>
                  <dt>Checksum</dt>
                  <dd>{platform.checksum}</dd>
                </div>
                <div>
                  <dt>Signing</dt>
                  <dd>{platform.signing}</dd>
                </div>
              </dl>

              <div className="download-action">
                <Button
                  className="w-full justify-between"
                  aria-disabled="true"
                  aria-describedby={statusId}
                >
                  <span className="flex items-center gap-2">
                    <ArrowDownToLine aria-hidden="true" size={18} />
                    {platform.label}
                  </span>
                  <span className="text-xs text-ink-950/62">Not wired</span>
                </Button>
              </div>
            </article>
          </Reveal>
        );
      })}
    </div>
  );
}

export function ReleaseReadiness() {
  return (
    <div className="release-readiness">
      <div className="release-readiness-head">
        <span className="release-info-icon">
          <Info aria-hidden="true" size={18} />
        </span>
        <div>
          <h3>Download status is intentionally honest.</h3>
          <p>
            Buttons are present for layout and product flow, but they do not start downloads until release artifacts are verified, signed, and checksummed.
          </p>
        </div>
      </div>
      <div className="release-state-grid">
        {releaseStates.map((item) => {
          const Icon = item.icon;
          return (
            <div key={item.label} className="release-state">
              <Icon aria-hidden="true" size={18} />
              <p>{item.label}</p>
              <span>{item.status}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
}
