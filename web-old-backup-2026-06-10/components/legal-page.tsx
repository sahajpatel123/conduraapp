import { legalUpdated } from "@/lib/site-data";

type Section = {
  id: string;
  title: string;
  body: string[];
};

export function LegalPage({
  title,
  lead,
  sections,
}: {
  title: string;
  lead: string;
  sections: Section[];
}) {
  return (
    <main id="main-content" className="container-prose py-16 sm:py-20">
      <p className="section-label">Last updated {legalUpdated}</p>
      <h1 className="page-title">{title}</h1>
      <p className="mt-6 text-lg leading-8 text-white/68">{lead}</p>

      <nav className="mt-10 rounded-lg border border-white/10 bg-white/[0.035] p-5" aria-label={`${title} table of contents`}>
        <h2 className="text-sm font-semibold text-white">Contents</h2>
        <ul className="mt-3 grid gap-2">
          {sections.map((section) => (
            <li key={section.id}>
              <a href={`#${section.id}`} className="text-sm text-white/62 underline-offset-4 hover:text-white hover:underline focus-ring">
                {section.title}
              </a>
            </li>
          ))}
        </ul>
      </nav>

      <div className="mt-12 space-y-12">
        {sections.map((section) => (
          <section key={section.id} id={section.id} className="scroll-mt-24">
            <h2 className="text-2xl font-semibold text-white">{section.title}</h2>
            <div className="mt-4 space-y-4">
              {section.body.map((paragraph) => (
                <p key={paragraph} className="leading-7 text-white/64">
                  {paragraph}
                </p>
              ))}
            </div>
          </section>
        ))}
      </div>
    </main>
  );
}
