"use client";
import { cn } from "@/lib/utils";

interface DocumentCardProps {
  title: string;
  url: string;
  content: string;
  createdAt: Date;
  previewImage?: string;
}

export function DocumentCard({
  title,
  url,
  content,
  createdAt,
  previewImage,
}: DocumentCardProps) {
  return (
    <div className="w-full group/card">
      <a
        href={url}
        target="_blank"
        rel="noopener noreferrer"
        className={cn(
          "cursor-pointer overflow-hidden relative card rounded-md shadow-lg mx-auto flex flex-col p-6 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700"
        )}
      >
        <div className="flex flex-row items-center space-x-4 mb-4">
          <div className="h-10 w-10 rounded-full border-2 border-gray-200 dark:border-gray-600 flex items-center justify-center">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
              className="h-6 w-6 text-gray-600 dark:text-gray-400"
            >
              <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z" />
              <polyline points="14 2 14 8 20 8" />
            </svg>
          </div>
          <div className="flex flex-col">
            <p className="font-normal text-base text-gray-600 dark:text-gray-300">
              {new URL(url).hostname}
            </p>
          </div>
        </div>

        <div className="text content">
          <h1 className="font-bold text-xl text-gray-900 dark:text-white mb-2">
            {title}
          </h1>
          <p className="font-normal text-gray-600 dark:text-gray-300 mb-4">
            {content}
          </p>
          <div className="text-sm text-gray-500 dark:text-gray-400">{url}</div>
        </div>
      </a>
    </div>
  );
}
