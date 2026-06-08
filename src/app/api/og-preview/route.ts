import { NextRequest, NextResponse } from 'next/server';

export async function GET(request: NextRequest) {
  const url = request.nextUrl.searchParams.get('url');

  if (!url) {
    return NextResponse.json({ error: 'URL manquante' }, { status: 400 });
  }

  try {
    const response = await fetch(url, {
      headers: {
        'User-Agent': 'Mozilla/5.0 (compatible; OtakuQuizBot/1.0)',
      },
      signal: AbortSignal.timeout(5000),
    });

    const html = await response.text();

    const getMetaContent = (property: string): string => {
      const regex = new RegExp(`<meta[^>]*(?:property|name)=["']${property}["'][^>]*content=["']([^"']+)["']`, 'i');
      const match = html.match(regex);
      return match ? match[1] : '';
    };

    const title = getMetaContent('og:title') || 
                  getMetaContent('twitter:title') || 
                  html.match(/<title[^>]*>([^<]+)<\/title>/i)?.[1] || '';

    const image = getMetaContent('og:image') || 
                  getMetaContent('twitter:image') || '';

    const domain = new URL(url).hostname;

    return NextResponse.json({
      title: title.substring(0, 200),
      image: image.substring(0, 500),
      domain,
    });
  } catch (error) {
    console.error('Erreur OG preview:', error);
    return NextResponse.json({
      title: '',
      image: '',
      domain: url ? new URL(url).hostname : '',
    });
  }
}