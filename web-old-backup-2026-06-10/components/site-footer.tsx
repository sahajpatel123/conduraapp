import Link from "next/link";
import { footerGroups } from "@/lib/site-data";

export function SiteFooter() {
  return (
    <footer className="site-footer" aria-label="Site footer">
      <div className="container-wide py-12">
        <div className="grid gap-10 md:grid-cols-[1.2fr_2fr]">
          <div>
            <div className="footer-brand">
              <span className="footer-mark">
                <span />
              </span>
              <span>Synaptic</span>
            </div>
            <p className="footer-copy">
              A free, local-first desktop AI agent that acts only through visible safety boundaries.
            </p>
            <p className="footer-meta">© 2026 Synaptic. Freeware binary. Proprietary source.</p>
          </div>

          <nav className="grid gap-8 sm:grid-cols-3" aria-label="Legal and support">
            {footerGroups.map((group) => (
              <div key={group.title}>
                <h2 className="footer-heading">{group.title}</h2>
                <ul className="footer-links">
                  {group.links.map((link) => {
                    const Icon = "icon" in link ? link.icon : undefined;
                    return (
                      <li key={link.href}>
                        <Link
                          href={link.href}
                          className="footer-link focus-ring"
                        >
                          {Icon ? <Icon aria-hidden="true" size={14} /> : null}
                          {link.label}
                        </Link>
                      </li>
                    );
                  })}
                </ul>
              </div>
            ))}
          </nav>
        </div>
      </div>
    </footer>
  );
}
