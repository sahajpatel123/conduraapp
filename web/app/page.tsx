"use client";

import Link from "next/link";
import { motion } from "motion/react";
import { SITE } from "@/lib/site";

const FEATURES = [
  {
    title: "Runs locally with Ollama",
    body: "Point Synaptic at your own models running on your machine. No cloud round-trips, no rented GPUs, no usage meter.",
  },
  {
    title: "Your data never leaves your machine",
    body: "Memory, API keys, skills, and audit logs are stored locally and encrypted at rest. Nothing is uploaded by default.",
  },
  {
    title: "Voice control",
    body: "Talk to your computer. Dictate, command, and steer the agent hands-free with low-latency local speech.",
  },
  {
    title: "Skills & automation",
    body: "Compose reusable skills that chain tools and actions, so repetitive work runs itself — under your supervision.",
  },
  {
    title: "End-to-end encrypted device sync",
    body: "Keep settings and memory in step across your devices. Synced data is encrypted so only your devices can read it.",
  },
  {
    title: "Open & auditable",
    body: "Every action is recorded in a tamper-resistant, append-only log. You can always see exactly what happened.",
  },
] as const;

const PLATFORM_BUTTONS = [
  { label: "Download for macOS", primary: true },
  { label: "Download for Windows", primary: false },
  { label: "Download for Linux", primary: false },
] as const;

const fadeUp = {
  hidden: { opacity: 0, y: 16 },
  visible: { opacity: 1, y: 0 },
} as const;

export default function Home() {
  return (
    <main className="mx-auto max-w-6xl px-6">
      <section className="flex flex-col items-center pt-24 pb-16 text-center sm:pt-32">
        <motion.span
          initial="hidden"
          animate="visible"
          variants={fadeUp}
          transition={{ duration: 0.4 }}
          className="rounded-full border border-neutral-800 px-3 py-1 text-xs font-medium text-neutral-400"
        >
          Privacy-first · Local-first · Free
        </motion.span>

        <motion.h1
          initial="hidden"
          animate="visible"
          variants={fadeUp}
          transition={{ duration: 0.5, delay: 0.05 }}
          className="mt-6 max-w-3xl text-balance text-5xl font-semibold tracking-tight text-white sm:text-6xl"
        >
          AI on your computer, free
        </motion.h1>

        <motion.p
          initial="hidden"
          animate="visible"
          variants={fadeUp}
          transition={{ duration: 0.5, delay: 0.1 }}
          className="mt-6 max-w-2xl text-lg text-neutral-400"
        >
          Synaptic is a local-first desktop AI agent that runs on your own
          models and your own machine. Your data stays with you — nothing leaves
          your computer unless you say so.
        </motion.p>

        <motion.div
          initial="hidden"
          animate="visible"
          variants={fadeUp}
          transition={{ duration: 0.5, delay: 0.15 }}
          className="mt-10 flex flex-wrap items-center justify-center gap-3"
        >
          {PLATFORM_BUTTONS.map((b) => (
            <Link
              key={b.label}
              href="/download"
              className={
                b.primary
                  ? "rounded-md bg-white px-5 py-2.5 text-sm font-medium text-black transition-colors hover:bg-neutral-200"
                  : "rounded-md border border-neutral-700 px-5 py-2.5 text-sm font-medium text-neutral-200 transition-colors hover:border-neutral-500 hover:text-white"
              }
            >
              {b.label}
            </Link>
          ))}
        </motion.div>
        <p className="mt-4 text-xs text-neutral-600">
          Free forever. No account required.
        </p>
      </section>

      <motion.section
        initial="hidden"
        whileInView="visible"
        viewport={{ once: true, amount: 0.2 }}
        variants={fadeUp}
        transition={{ duration: 0.5 }}
        className="pb-20"
      >
        <div className="relative overflow-hidden rounded-2xl border border-neutral-800 bg-gradient-to-b from-neutral-900 to-neutral-950 p-2 shadow-2xl shadow-black/40">
          <div className="flex aspect-video w-full items-center justify-center rounded-xl bg-[radial-gradient(circle_at_50%_0%,rgba(255,255,255,0.06),transparent_60%)]">
            <div className="flex flex-col items-center gap-3 text-center">
              <div className="flex gap-1.5">
                <span className="h-3 w-3 rounded-full bg-neutral-700" />
                <span className="h-3 w-3 rounded-full bg-neutral-700" />
                <span className="h-3 w-3 rounded-full bg-neutral-700" />
              </div>
              <p className="text-sm font-medium text-neutral-400">
                Press {SITE.name === "Synaptic" ? "Cmd+Shift+Space" : "your hotkey"} to summon
              </p>
              <p className="max-w-sm text-xs text-neutral-600">
                Overlay appears. Type or speak. Agent responds. Your data stays local.
              </p>
            </div>
          </div>
        </div>
      </motion.section>

      <section className="border-t border-neutral-900 py-20">
        <div className="mx-auto max-w-2xl text-center">
          <h2 className="text-3xl font-semibold tracking-tight text-white">
            Built for people who want control
          </h2>
          <p className="mt-4 text-neutral-400">
            Everything runs where you can see it. Capable by default, private by
            design.
          </p>
        </div>

        <div className="mt-14 grid gap-px overflow-hidden rounded-2xl border border-neutral-800 bg-neutral-800 sm:grid-cols-2 lg:grid-cols-3">
          {FEATURES.map((f, i) => (
            <motion.div
              key={f.title}
              initial="hidden"
              whileInView="visible"
              viewport={{ once: true, amount: 0.3 }}
              variants={fadeUp}
              transition={{ duration: 0.4, delay: (i % 3) * 0.05 }}
              className="bg-neutral-950 p-7"
            >
              <h3 className="text-base font-medium text-white">{f.title}</h3>
              <p className="mt-2 text-sm leading-relaxed text-neutral-400">
                {f.body}
              </p>
            </motion.div>
          ))}
        </div>
      </section>

      <section className="border-t border-neutral-900 py-20">
        <div className="rounded-2xl border border-neutral-800 bg-gradient-to-b from-neutral-900/60 to-neutral-950 px-8 py-14 text-center">
          <h2 className="text-3xl font-semibold tracking-tight text-white">
            Ready when you are
          </h2>
          <p className="mx-auto mt-4 max-w-xl text-neutral-400">
            {SITE.description}
          </p>
          <Link
            href="/download"
            className="mt-8 inline-block rounded-md bg-white px-6 py-3 text-sm font-medium text-black transition-colors hover:bg-neutral-200"
          >
            Download Synaptic
          </Link>
        </div>
      </section>
    </main>
  );
}
