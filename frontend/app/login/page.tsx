"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { motion } from "framer-motion";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { loginRequest } from "@/lib/auth";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";

const schema = z.object({
  username: z.string().min(1, "Username is required"),
  password: z.string().min(1, "Password is required"),
});

type FormData = z.infer<typeof schema>;

export default function LoginPage() {
  const router = useRouter();
  const [serverError, setServerError] = useState("");
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormData>({ resolver: zodResolver(schema) });

  const onSubmit = async (data: FormData) => {
    setServerError("");
    try {
      await loginRequest(data.username, data.password);
      router.push("/vinyls");
    } catch (err) {
      setServerError(err instanceof Error ? err.message : "Login failed");
    }
  };

  return (
    <main className="min-h-screen flex items-center justify-center bg-background px-4">
      <div
        className="absolute inset-0 opacity-[0.03] pointer-events-none"
        style={{
          backgroundImage: `repeating-radial-gradient(circle at 50% 50%, transparent 0, transparent 20px, #D4A017 20px, #D4A017 21px)`,
        }}
      />

      <motion.div
        initial={{ opacity: 0, y: 24 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: "easeOut" }}
        className="relative w-full max-w-sm"
      >
        <div className="text-center mb-10">
          <div className="w-16 h-16 rounded-full bg-accent/10 border-2 border-accent/30 flex items-center justify-center mx-auto mb-4">
            <span className="text-2xl">◉</span>
          </div>
          <h1 className="font-serif text-3xl font-bold text-foreground">Vinyl Catalog</h1>
          <p className="text-muted text-sm mt-1">Sign in to your collection</p>
        </div>

        <form
          onSubmit={handleSubmit(onSubmit)}
          className="bg-surface border border-border rounded-xl p-8 flex flex-col gap-5"
        >
          <Input
            id="username"
            label="Username"
            placeholder="admin"
            autoComplete="username"
            {...register("username")}
            error={errors.username?.message}
          />
          <Input
            id="password"
            label="Password"
            type="password"
            placeholder="••••••••"
            autoComplete="current-password"
            {...register("password")}
            error={errors.password?.message}
          />

          {serverError && (
            <p className="text-sm text-red-400 text-center">{serverError}</p>
          )}

          <Button type="submit" loading={isSubmitting} className="w-full mt-2">
            Sign in
          </Button>
        </form>
      </motion.div>
    </main>
  );
}
