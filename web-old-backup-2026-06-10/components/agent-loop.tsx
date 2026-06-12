import { agentLoop } from "@/lib/site-data";
import { Reveal } from "@/components/reveal";

export function AgentLoop() {
  return (
    <div className="approval-path">
      {agentLoop.map((step, index) => {
        const Icon = step.icon;
        return (
          <Reveal key={step.title} delay={index * 0.07}>
            <article className="approval-step" tabIndex={0}>
              <div className="loop-step-top">
                <span>0{index + 1}</span>
                <Icon aria-hidden="true" size={20} />
              </div>
              <h3>{step.title}</h3>
              <p>{step.body}</p>
            </article>
          </Reveal>
        );
      })}
    </div>
  );
}
