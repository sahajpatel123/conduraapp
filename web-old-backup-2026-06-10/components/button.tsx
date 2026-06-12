import Link from "next/link";
import type { AnchorHTMLAttributes, ButtonHTMLAttributes, ReactNode } from "react";

type Variant = "primary" | "secondary" | "ghost" | "danger";

const variantClass: Record<Variant, string> = {
  primary:
    "border-slate-950 bg-slate-950 text-white hover:bg-[#263027] focus-visible:outline-slate-950",
  secondary:
    "border-black/15 bg-white/55 text-slate-950 hover:border-slate-950/40 hover:bg-white focus-visible:outline-slate-950",
  ghost:
    "border-transparent bg-transparent text-slate-600 hover:text-slate-950 hover:bg-black/[0.045] focus-visible:outline-slate-950",
  danger:
    "border-red-900/25 bg-red-900/8 text-red-900 hover:bg-red-900/12 focus-visible:outline-red-900",
};

const base =
  "button-like inline-flex min-h-11 items-center justify-center gap-2 rounded-full border px-5 py-2.5 text-sm font-semibold transition duration-200 ease-out hover:-translate-y-0.5 active:translate-y-0 active:scale-[0.99] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2";

type LinkButtonProps = AnchorHTMLAttributes<HTMLAnchorElement> & {
  href: string;
  variant?: Variant;
  children: ReactNode;
};

type NativeButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: Variant;
  children: ReactNode;
};

export function LinkButton({ href, variant = "primary", className = "", children, ...props }: LinkButtonProps) {
  return (
    <Link href={href} className={`${base} button-${variant} ${variantClass[variant]} ${className}`} {...props}>
      {children}
    </Link>
  );
}

export function Button({ variant = "primary", className = "", children, ...props }: NativeButtonProps) {
  return (
    <button className={`${base} button-${variant} ${variantClass[variant]} ${className}`} type="button" {...props}>
      {children}
    </button>
  );
}
