"use client";

import { motion } from "framer-motion";

interface Props {
  children: React.ReactNode;
  type?: "button" | "submit" | "reset";
  loading?: boolean;
  className?: string;
  onClick?: () => void;
  variant?: "default" | "ghost" | "danger";
}

export function Button({ children, type = "button", loading = false, className = "", onClick, variant = "default" }: Props) {
  const baseStyles = "px-4 py-2 rounded font-medium text-sm disabled:opacity-50 disabled:cursor-not-allowed";
  
  const variantStyles = {
    default: "bg-accent text-background",
    ghost: "bg-transparent text-foreground hover:bg-surface",
    danger: "bg-red-600 text-white hover:bg-red-700",
  };

  return (
    <motion.button
      type={type}
      onClick={onClick}
      disabled={loading}
      whileHover={{ scale: loading ? 1 : 1.02 }}
      whileTap={{ scale: loading ? 1 : 0.98 }}
      className={`${baseStyles} ${variantStyles[variant]} ${className}`}
    >
      {loading ? "Loading..." : children}
    </motion.button>
  );
}
