import type { Metadata } from "next";
import { GeistSans } from "geist/font/sans";
import { AuthProvider } from "@/components/providers/AuthProvider";
import { Toaster } from "@/components/ui/Toaster";
import "./globals.css";

export const metadata: Metadata = {
  title: 'Otaku Quiz Africa - Le Quiz Anime de l Afrique',
  description: 'Affronte les meilleurs fans d anime d Afrique francophone. Quiz compétitifs, classements live, et milliers de quiz créés par la communauté.',
  keywords: ['anime', 'quiz', 'afrique', 'francophone', 'manga', 'otaku', 'jeu'],
  authors: [{ name: 'Otaku Quiz Africa' }],
  openGraph: {
    title: 'Otaku Quiz Africa',
    description: 'Le quiz anime de l Afrique francophone',
    type: 'website',
    locale: 'fr_FR',
    images: ['/og-image.jpg'],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Otaku Quiz Africa',
    description: 'Le quiz anime de l Afrique francophone',
    images: ['/og-image.jpg'],
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="fr" className="dark">
      <body className={`${GeistSans.variable} font-sans antialiased bg-dark min-h-screen`}>
        <AuthProvider>
          {children}
          <Toaster />
        </AuthProvider>
      </body>
    </html>
  );
}

