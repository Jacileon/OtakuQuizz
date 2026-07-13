import { MetadataRoute } from '../../node_modules/next';
import { createClient } from '@/lib/supabase/server';

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const supabase = createClient();

  const { data: quizzes } = await supabase
    .from('quizzes')
    .select('id, updated_at')
    .eq('is_visible', true)
    .limit(100);

  const quizUrls = (quizzes || []).map((quiz) => ({
    url: `${process.env.NEXT_PUBLIC_APP_URL}/quiz/${quiz.id}`,
    lastModified: quiz.updated_at,
    changeFrequency: 'weekly' as const,
    priority: 0.8,
  }));

  return [
    { url: process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000', lastModified: new Date(), changeFrequency: 'daily', priority: 1 },
    { url: `${process.env.NEXT_PUBLIC_APP_URL}/explore`, lastModified: new Date(), changeFrequency: 'daily', priority: 0.9 },
    { url: `${process.env.NEXT_PUBLIC_APP_URL}/leaderboard`, lastModified: new Date(), changeFrequency: 'daily', priority: 0.9 },
    ...quizUrls,
  ];
}

