// ============================================================
// LANDING PAGE - Page d'accueil publique
// ============================================================

import Link from '../../node_modules/next/link';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import {
  Sword,
  Trophy,
  Users,
  Zap,
  ChevronRight,
  Globe,
  Gamepad2,
  BarChart3,
} from 'lucide-react';
import { createClient } from '@/lib/supabase/server';

export default async function LandingPage() {
  const supabase = createClient();

  // Stats pour le hero
  const { count: quizCount } = await supabase
    .from('quizzes')
    .select('*', { count: 'exact', head: true })
    .eq('is_visible', true);

  const { count: playerCount } = await supabase
    .from('user_profiles')
    .select('*', { count: 'exact', head: true });

  const { data: featuredQuizzes } = await supabase
    .from('quizzes')
    .select('*, creator:creator_id(username, avatar_url)')
    .eq('is_visible', true)
    .order('play_count', { ascending: false })
    .limit(6);

  const { data: topPlayers } = await supabase
    .rpc('get_global_leaderboard', { limit_count: 5 });

  return (
    <div className="min-h-screen bg-dark">
      {/* HERO SECTION */}
      <section className="relative overflow-hidden py-20 sm:py-32">
        <div className="absolute inset-0 bg-gradient-to-br from-brand/10 via-transparent to-accent/10" />
        <div className="absolute inset-0 opacity-20">
          <div className="absolute top-20 left-10 w-72 h-72 bg-brand/20 rounded-full blur-3xl" />
          <div className="absolute bottom-20 right-10 w-96 h-96 bg-accent/20 rounded-full blur-3xl" />
        </div>

        <div className="container relative z-10 mx-auto px-4">
          <div className="flex flex-col lg:flex-row items-center gap-12">
            <div className="flex-1 text-center lg:text-left">
              <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-brand/10 border border-brand/20 text-brand text-sm mb-6">
                <Zap className="h-4 w-4" />
                <span>La communauté otaku africaine</span>
              </div>

              <h1 className="font-display text-5xl sm:text-7xl lg:text-8xl tracking-wider text-white mb-6">
                LE QUIZ ANIME
                <span className="block text-brand">DE L'AFRIQUE</span>
              </h1>

              <p className="text-lg text-muted-foreground max-w-xl mx-auto lg:mx-0 mb-8">
                Affronte les meilleurs fans d'anime d'Afrique francophone. 
                Des milliers de quiz, des événements compétitifs et un classement en temps réel.
              </p>

              {/* Compteurs */}
              <div className="flex flex-wrap justify-center lg:justify-start gap-8 mb-10">
                <div className="text-center">
                  <div className="font-display text-3xl text-brand">{quizCount?.toLocaleString() || '0'}</div>
                  <div className="text-sm text-muted-foreground">Quiz disponibles</div>
                </div>
                <div className="text-center">
                  <div className="font-display text-3xl text-accent">{playerCount?.toLocaleString() || '0'}</div>
                  <div className="text-sm text-muted-foreground">Joueurs actifs</div>
                </div>
                <div className="text-center">
                  <div className="font-display text-3xl text-white">20+</div>
                  <div className="text-sm text-muted-foreground">Pays</div>
                </div>
              </div>

              <div className="flex flex-col sm:flex-row gap-4 justify-center lg:justify-start">
                <Link href="/login">
                  <Button size="lg" className="gap-2 text-lg px-8">
                    <Sword className="h-5 w-5" />
                    Rejoindre la communauté
                  </Button>
                </Link>
                <Link href="/explore">
                  <Button size="lg" variant="outline" className="gap-2 text-lg px-8">
                    Explorer les quiz
                    <ChevronRight className="h-5 w-5" />
                  </Button>
                </Link>
              </div>
            </div>

            {/* Preview Mockup */}
            <div className="flex-1 hidden lg:block">
              <div className="relative">
                <div className="absolute -inset-4 bg-gradient-to-r from-brand/20 to-accent/20 rounded-2xl blur-xl" />
                <Card className="relative border-dark-border bg-dark-card/80 backdrop-blur">
                  <CardContent className="p-6 space-y-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-full bg-brand/20 flex items-center justify-center">
                          <Sword className="h-5 w-5 text-brand" />
                        </div>
                        <div>
                          <div className="font-medium">Quiz Naruto</div>
                          <div className="text-xs text-muted-foreground">25 questions</div>
                        </div>
                      </div>
                      <div className="text-sm text-accent font-medium">En cours</div>
                    </div>
                    <div className="h-2 bg-dark-surface rounded-full overflow-hidden">
                      <div className="h-full w-3/4 bg-brand rounded-full" />
                    </div>
                    <div className="grid grid-cols-2 gap-2">
                      <div className="p-3 rounded-lg bg-dark-surface border border-dark-border text-sm">Rasengan</div>
                      <div className="p-3 rounded-lg bg-dark-surface border border-dark-border text-sm">Chidori</div>
                      <div className="p-3 rounded-lg bg-brand/20 border border-brand/30 text-sm">Kamehameha</div>
                      <div className="p-3 rounded-lg bg-dark-surface border border-dark-border text-sm">Bankai</div>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* FEATURES */}
      <section className="py-20 bg-dark-surface/30">
        <div className="container mx-auto px-4">
          <h2 className="font-display text-3xl sm:text-4xl text-center mb-12 tracking-wider">
            POURQUOI <span className="text-brand">OTAKU QUIZ AFRICA</span> ?
          </h2>
          <div className="grid md:grid-cols-3 gap-8">
            <FeatureCard
              icon={<Trophy className="h-8 w-8 text-brand" />}
              title="Quiz Compétitifs"
              description="Des événements officiels chaque semaine avec des récompenses exclusives et un classement en temps réel."
            />
            <FeatureCard
              icon={<BarChart3 className="h-8 w-8 text-accent" />}
              title="Classements Live"
              description="Affronte les meilleurs en temps réel. Suis ta progression et grimpe dans les rangs de F à Légende."
            />
            <FeatureCard
              icon={<Users className="h-8 w-8 text-green-400" />}
              title="Créé par des fans"
              description="Des milliers de quiz créés par la communauté africaine. Anime, manga, openings et plus encore."
            />
          </div>
        </div>
      </section>

      {/* FEATURED QUIZZES */}
      <section className="py-20">
        <div className="container mx-auto px-4">
          <div className="flex items-center justify-between mb-8">
            <h2 className="font-display text-3xl tracking-wider">
              QUIZ EN <span className="text-brand">VEDETTE</span>
            </h2>
            <Link href="/explore">
              <Button variant="ghost" className="gap-2">
                Voir tous <ChevronRight className="h-4 w-4" />
              </Button>
            </Link>
          </div>
          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {featuredQuizzes?.map((quiz) => (
              <QuizPreviewCard key={quiz.id} quiz={quiz} />
            )) || Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="h-48 rounded-lg bg-dark-surface animate-pulse" />
            ))}
          </div>
        </div>
      </section>

      {/* LEADERBOARD PREVIEW */}
      <section className="py-20 bg-dark-surface/30">
        <div className="container mx-auto px-4">
          <h2 className="font-display text-3xl tracking-wider text-center mb-8">
            CLASSEMENT <span className="text-brand">MENSUEL</span>
          </h2>
          <div className="max-w-2xl mx-auto space-y-3">
            {topPlayers?.map((player: any, i: number) => (
              <div
                key={player.user_id}
                className="flex items-center gap-4 p-4 rounded-lg bg-dark-card border border-dark-border hover:border-brand/30 transition-colors"
              >
                <div className={
                  i === 0 ? 'text-2xl' : i === 1 ? 'text-xl' : i === 2 ? 'text-lg' : 'text-base'
                }>
                  {i === 0 ? '🥇' : i === 1 ? '🥈' : i === 2 ? '🥉' : `#${i + 1}`}
                </div>
                <div className="flex-1">
                  <div className="font-medium">{player.username}</div>
                  <div className="text-xs text-muted-foreground">{player.quizzes_played} quiz joués</div>
                </div>
                <div className="font-display text-xl text-brand">{player.xp?.toLocaleString()} XP</div>
              </div>
            )) || Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="h-16 rounded-lg bg-dark-surface animate-pulse" />
            ))}
          </div>
          <div className="text-center mt-8">
            <Link href="/leaderboard">
              <Button variant="outline" size="lg">Voir le classement complet</Button>
            </Link>
          </div>
        </div>
      </section>

      {/* CTA FINAL */}
      <section className="py-20">
        <div className="container mx-auto px-4 text-center">
          <h2 className="font-display text-4xl sm:text-5xl mb-6 tracking-wider">
            PRÊT À <span className="text-brand">PROUVER</span> QUE TU ES UN VRAI OTAKU ?
          </h2>
          <p className="text-muted-foreground mb-8 max-w-xl mx-auto">
            Rejoins des milliers de fans d'anime à travers l'Afrique et montre tes connaissances.
          </p>
          <Link href="/login">
            <Button size="lg" className="text-lg px-10 py-6">
              <Globe className="h-5 w-5 mr-2" />
              Commencer maintenant
            </Button>
          </Link>
        </div>
      </section>

      {/* FOOTER */}
      <footer className="border-t border-dark-border py-12">
        <div className="container mx-auto px-4">
          <div className="flex flex-col md:flex-row items-center justify-between gap-6">
            <div className="flex items-center gap-2">
              <Sword className="h-6 w-6 text-brand" />
              <span className="font-display text-lg tracking-wider">OTAKU QUIZ AFRICA</span>
            </div>
            <div className="flex gap-6 text-sm text-muted-foreground">
              <Link href="/about" className="hover:text-white transition-colors">À propos</Link>
              <Link href="/contact" className="hover:text-white transition-colors">Contact</Link>
              <Link href="/terms" className="hover:text-white transition-colors">Conditions</Link>
              <Link href="/privacy" className="hover:text-white transition-colors">Confidentialité</Link>
            </div>
            <div className="text-sm text-muted-foreground">
              Fait avec ❤️ pour les fans d'anime d'Afrique
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}

function FeatureCard({ icon, title, description }: { icon: React.ReactNode; title: string; description: string }) {
  return (
    <Card className="border-dark-border bg-dark-card/50 hover:border-brand/30 transition-all duration-300 hover:-translate-y-1">
      <CardContent className="p-6">
        <div className="mb-4">{icon}</div>
        <h3 className="font-display text-xl mb-2 tracking-wider">{title}</h3>
        <p className="text-sm text-muted-foreground">{description}</p>
      </CardContent>
    </Card>
  );
}

function QuizPreviewCard({ quiz }: { quiz: any }) {
  return (
    <Card className="border-dark-border bg-dark-card/50 hover:border-brand/30 transition-all duration-300 overflow-hidden group">
      <div className="h-32 bg-gradient-to-br from-brand/20 to-accent/20 flex items-center justify-center relative overflow-hidden">
        <Gamepad2 className="h-12 w-12 text-white/20 group-hover:scale-110 transition-transform" />
        <div className="absolute bottom-2 right-2">
          <span className="px-2 py-1 text-xs rounded-md bg-dark/80 text-white">
            {quiz.question_count} questions
          </span>
        </div>
      </div>
      <CardContent className="p-4">
        <h3 className="font-medium mb-1 truncate">{quiz.title}</h3>
        <div className="flex items-center justify-between text-xs text-muted-foreground">
          <span>{quiz.series}</span>
          <span>{quiz.play_count} plays</span>
        </div>
        {quiz.creator && (
          <div className="flex items-center gap-2 mt-3 pt-3 border-t border-dark-border">
            <div className="h-5 w-5 rounded-full bg-dark-surface" />
            <span className="text-xs">{quiz.creator.username}</span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

