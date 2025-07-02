import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import AuthProvider from "@/providers/AuthProvider";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Foundry - Store & Find Information When You Need It",
  description:
    "Foundry is a powerful tool for storing and finding information when you actually need it. Clean your tabs without losing any information. The best way to organize your digital workspace.",
  keywords: [
    "information management",
    "tab organization",
    "productivity tool",
    "digital workspace",
    "information storage",
    "search tool",
    "browser tabs",
    "knowledge management",
    "workflow optimization",
    "digital organization",
    "tab cleaner",
  ],
  authors: [{ name: "Foundry Team" }],
  creator: "Foundry",
  publisher: "Foundry",
  formatDetection: {
    email: false,
    address: false,
    telephone: false,
  },
  metadataBase: new URL("https://foundrymemory.vercel.app"),
  alternates: {
    canonical: "/",
  },
  openGraph: {
    title: "Foundry - Store & Find Information When You Need It",
    description:
      "Foundry is a powerful tool for storing and finding information when you actually need it. Clean your tabs without losing any information.",
    url: "https://foundrymemory.vercel.app",
    siteName: "Foundry",
    images: [
      {
        url: "/os.png",
        width: 1200,
        height: 630,
        alt: "Foundry - Information Management Tool",
      },
    ],
    locale: "en_US",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "Foundry - Store & Find Information When You Need It",
    description:
      "Foundry is a powerful tool for storing and finding information when you actually need it. Clean your tabs without losing any information.",
    images: ["/os.png"],
    creator: "@foundryapp",
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      "max-video-preview": -1,
      "max-image-preview": "large",
      "max-snippet": -1,
    },
  },
  category: "productivity",
  classification: "productivity tool",
  referrer: "origin-when-cross-origin",
  other: {
    "geo.region": "US",
    "geo.placename": "United States",
    "geo.position": "37.7749;-122.4194",
    ICBM: "37.7749, -122.4194",
    "DC.title": "Foundry - Information Management Tool",
    "DC.creator": "Foundry Team",
    "DC.subject": "Productivity, Information Management, Tab Organization",
    "DC.description":
      "Foundry is a powerful tool for storing and finding information when you actually need it.",
    "DC.publisher": "Foundry",
    "DC.contributor": "Foundry Team",
    "DC.date": "2024",
    "DC.type": "Software",
    "DC.format": "Web Application",
    "DC.identifier": "https://foundrymemory.vercel.app",
    "DC.language": "en",
    "DC.coverage": "Worldwide",
    "DC.rights": "Copyright Foundry 2024",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
        suppressHydrationWarning
      >
        <AuthProvider>{children}</AuthProvider>
      </body>
    </html>
  );
}
