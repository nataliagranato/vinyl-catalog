"use client";

import { useRouter } from "next/navigation";
import Link from "next/link";
import { motion } from "framer-motion";
import { ArrowLeft } from "lucide-react";
import { VinylForm, VinylFormData } from "@/components/VinylForm";
import { useToast } from "@/components/ToastProvider";

export default function NewVinylPage() {
  const router = useRouter();
  const { toast } = useToast();

  const handleSubmit = async (data: VinylFormData) => {
    const res = await fetch("/api/vinyls", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error ?? "Failed to create vinyl");
    }
    toast("Vinyl added to collection", "success");
    router.push("/vinyls");
  };

  return (
    <div className="min-h-screen bg-background px-4 py-8">
      <div className="max-w-md mx-auto">
        <Link href="/vinyls" className="inline-flex items-center gap-2 text-muted hover:text-accent transition-colors text-sm mb-8">
          <ArrowLeft size={14} /> Back to collection
        </Link>

        <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }}>
          <h1 className="font-serif text-3xl font-bold mb-2">Add a Record</h1>
          <p className="text-muted text-sm mb-8">Add a new vinyl to your collection</p>

          <div className="bg-surface border border-border rounded-xl p-8">
            <VinylForm onSubmit={handleSubmit} submitLabel="Add to collection" />
          </div>
        </motion.div>
      </div>
    </div>
  );
}
