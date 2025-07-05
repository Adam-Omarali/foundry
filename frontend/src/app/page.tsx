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
  const { jwt } = useAuth();
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<any[]>([]);

  const sortedResults = (results: any[]) => {
    return results.sort((a, b) => {
      if (a.read !== b.read) {
        return a.read ? 1 : -1;
      }
      return 0;
    });
  };

  const handleKeyPress = async (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" && query.trim() && jwt) {
      try {
        const response = await fetch(
          `${process.env.NEXT_PUBLIC_API_URL}/api/documents`,
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              Authorization: `Bearer ${jwt}`,
            },
            body: JSON.stringify({ query: query.trim() }),
          }
        );

        if (!response.ok) {
          throw new Error("Failed to fetch documents");
        }

        const data = await response.json();
        console.log(data);
        setResults(data);
      } catch (error) {
        console.error("Error fetching documents:", error);
      }
    }
  };

  const handleMarkAsRead = async (documentId: number) => {
    if (!jwt) return;

    try {
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/documents/read`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${jwt}`,
          },
          body: JSON.stringify({ document_id: documentId }),
        }
      );

      if (!response.ok) {
        throw new Error("Failed to mark document as read");
      }

      // Update the local state to reflect the change
      setResults((prevResults) =>
        prevResults.map((doc: any) =>
          doc.id === documentId ? { ...doc, read: true } : doc
        )
      );
    } catch (error) {
      console.error("Error marking document as read:", error);
    }
  };

  return (
    <div className="flex flex-col items-center min-h-screen w-[90%] mx-auto py-8">
      <LabelInputContainer className="max-w-2xl w-full mb-8">
        <Input
          id="search"
          placeholder=""
          type="text"
          autoComplete="off"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyUp={handleKeyPress}
        />
      </LabelInputContainer>

      {results && results.length > 0 && (
        <div className="space-y-4 w-[90%] mx-auto">
          {sortedResults(results).map((doc: any) => (
            <DocumentCard
              key={doc.id}
              id={doc.id}
              title={doc.title}
              url={doc.url}
              content={doc.content}
              read={doc.read}
              onMarkAsRead={handleMarkAsRead}
            />
          ))}
        </div>
      )}
    </div>
  );
}
