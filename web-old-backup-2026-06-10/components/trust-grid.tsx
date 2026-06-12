import type { LucideIcon } from "lucide-react";
import { Reveal } from "@/components/reveal";

type Item = {
  title: string;
  body: string;
  icon: LucideIcon;
};

export function TrustGrid({ items }: { items: Item[] }) {
  return (
    <div className={`trust-ledger-grid count-${items.length}`}>
      {items.map((item, index) => {
        const Icon = item.icon;
        return (
          <Reveal key={item.title} delay={index * 0.05}>
            <article className="trust-ledger-row">
              <div className="trust-icon">
                <Icon aria-hidden="true" size={21} />
              </div>
              <h3>{item.title}</h3>
              <p>{item.body}</p>
            </article>
          </Reveal>
        );
      })}
    </div>
  );
}
