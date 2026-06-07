import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",
  {
    variants: {
      variant: {
        default: "border-transparent bg-brand text-white hover:bg-brand-dark",
        secondary: "border-transparent bg-dark-surface text-white hover:bg-dark-card",
        destructive: "border-transparent bg-red-600 text-white hover:bg-red-700",
        outline: "text-white border-dark-border",
        accent: "border-transparent bg-accent text-white hover:bg-accent-dark",
        ghost: "border-transparent bg-transparent text-white hover:bg-dark-surface",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props} />
  );
}

export { Badge, badgeVariants };

