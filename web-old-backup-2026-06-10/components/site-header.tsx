"use client";

import Link from "next/link";
import { Menu, X } from "lucide-react";
import { useState } from "react";
import { primaryNav } from "@/lib/nav";
import { LinkButton } from "@/components/button";

export function SiteHeader() {
  const [open, setOpen] = useState(false);

  return (
    <header className="site-header">
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded-md focus:bg-slate-950 focus:px-4 focus:py-2 focus:text-white"
      >
        Skip to main content
      </a>
      <div className="site-header-inner container-wide">
        <Link href="/" className="brand focus-ring">
          <span className="brand-mark">
            <span className="brand-dot" />
          </span>
          <span className="brand-text">Synaptic</span>
        </Link>

        <nav className="site-nav" aria-label="Primary">
          {primaryNav.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className="focus-ring"
            >
              {item.label}
            </Link>
          ))}
        </nav>

        <div className="site-actions">
          <Link
            href="/dashboard"
            className="signin-link focus-ring"
          >
            Sign in
          </Link>
          <LinkButton href="/download" className="download-link min-h-10 px-4 py-2">
            Download
          </LinkButton>
        </div>

        <button
          type="button"
          className="mobile-menu-button focus-ring"
          aria-label={open ? "Close navigation" : "Open navigation"}
          aria-expanded={open}
          onClick={() => setOpen((value) => !value)}
        >
          {open ? <X aria-hidden="true" size={20} /> : <Menu aria-hidden="true" size={20} />}
        </button>
      </div>

      {open ? (
        <div className="mobile-nav">
          <nav className="mobile-nav-inner container-wide" aria-label="Mobile primary">
            {primaryNav.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className="focus-ring"
                onClick={() => setOpen(false)}
              >
                {item.label}
              </Link>
            ))}
            <div className="mobile-actions">
              <LinkButton href="/download" onClick={() => setOpen(false)}>
                Download
              </LinkButton>
              <LinkButton href="/dashboard" variant="secondary" onClick={() => setOpen(false)}>
                Sign in for dashboard
              </LinkButton>
            </div>
          </nav>
        </div>
      ) : null}
    </header>
  );
}
