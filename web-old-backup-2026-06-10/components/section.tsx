import type { ReactNode } from "react";

type SectionProps = {
  id?: string;
  children: ReactNode;
  className?: string;
  bleed?: boolean;
};

export function Section({ id, children, className = "", bleed = false }: SectionProps) {
  return (
    <section id={id} className={`section ${className}`}>
      <div className={bleed ? "mx-auto w-full" : "container-wide"}>{children}</div>
    </section>
  );
}

type SectionHeaderProps = {
  label?: string;
  title: string;
  lead?: string;
  align?: "left" | "center";
};

export function SectionHeader({ label, title, lead, align = "left" }: SectionHeaderProps) {
  return (
    <div className={align === "center" ? "mx-auto max-w-3xl text-center" : "max-w-3xl"}>
      {label ? <p className="section-label">{label}</p> : null}
      <h2 className="section-title">{title}</h2>
      {lead ? <p className="section-lead">{lead}</p> : null}
    </div>
  );
}
