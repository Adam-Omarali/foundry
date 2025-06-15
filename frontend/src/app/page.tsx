"use client";
import { useAuth } from "@/hooks/useAuth";
import { Input } from "@/components/ui/input";
import { DocumentCard } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { useState } from "react";

const LabelInputContainer = ({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) => {
  return (
    <div className={cn("flex w-full flex-col space-y-2", className)}>
      {children}
    </div>
  );
};

export default function Home() {
  const { accessToken, jwt, isAuthenticated, isLoading, user } = useAuth();
  const [query, setQuery] = useState("");
  const [results, setResults] = useState([]);

  const handleKeyPress = async (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" && query.trim() && jwt) {
      try {
        const response = await fetch("http://localhost:8080/api/documents", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${jwt}`,
          },
          body: JSON.stringify({ query: query.trim() }),
        });

        if (!response.ok) {
          throw new Error("Failed to fetch documents");
        }

        const data = await response.json();
        setResults(data);
      } catch (error) {
        console.error("Error fetching documents:", error);
      }
    }
  };

  return (
    <div className="flex flex-col items-center min-h-screen w-[90%] mx-auto py-8">
      <LabelInputContainer className="max-w-2xl w-full mb-8">
        <Input
          id="search"
          placeholder="Search your documents..."
          type="text"
          autoComplete="off"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyPress={handleKeyPress}
        />
      </LabelInputContainer>

      {results.length > 0 && (
        <div className="space-y-4 w-[90%] mx-auto">
          {results.map((doc: any) => (
            <DocumentCard
              key={doc.id}
              title={doc.title}
              url={doc.url}
              content={doc.content}
              createdAt={doc.created_at}
            />
          ))}
        </div>
      )}
    </div>
  );
}
