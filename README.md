# Otaku Quiz Africa

Plateforme web communautaire de quiz anime destinée à l'Afrique francophone.

## Stack Technique

- **Frontend**: Next.js 14 (App Router), TypeScript strict, Tailwind CSS
- **Backend**: Supabase (PostgreSQL + Auth + Realtime + Storage)
- **Médias**: Cloudinary (images, audio, GIFs)
- **Auth**: Google OAuth via Supabase Auth
- **Déploiement**: Vercel

## Installation

```bash
# 1. Cloner le projet
git clone <repo>
cd otaku-quiz-africa

# 2. Installer les dépendances
npm install

# 3. Configurer les variables d'environnement
cp .env.local.example .env.local
# Remplir avec tes clés Supabase et Cloudinary

# 4. Lancer en développement
npm run dev
```

## Configuration Supabase

1. Créer un projet sur [Supabase](https://supabase.com)
2. Exécuter la migration: `supabase/migrations/001_initial_schema.sql`
3. Exécuter les seeds: `supabase/seeds/badges.sql`
4. Configurer l'Auth Google OAuth dans le dashboard Supabase
5. Activer Realtime sur les tables nécessaires

## Configuration Cloudinary

1. Créer un compte sur [Cloudinary](https://cloudinary.com)
2. Récupérer le cloud name, API key et API secret
3. Remplir dans `.env.local`

## Variables d'Environnement

```env
NEXT_PUBLIC_SUPABASE_URL=your_supabase_url
NEXT_PUBLIC_SUPABASE_ANON_KEY=your_supabase_anon_key
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key
NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

## Structure des Dossiers

```
src/
  app/              # Routes Next.js (App Router)
  components/       # Composants React
    ui/            # Composants UI (shadcn)
    layout/        # Navbar, Sidebar, MobileNav
    dashboard/     # Composants dashboard
    quiz/          # Moteur de quiz
    quiz-creator/  # Création de quiz
    explore/       # Exploration
    profile/       # Profil
    leaderboard/   # Classements
    events/        # Événements
    notifications/ # Notifications
  lib/
    actions/       # Server Actions
    queries/       # Requêtes Supabase
    auth/          # Auth helpers
    supabase/      # Clients Supabase
    realtime/      # Subscriptions Realtime
    badges/        # Moteur de badges
    hooks/         # Hooks React
  types/           # Types TypeScript
supabase/
  migrations/      # Migrations SQL
  seeds/           # Données initiales
```

## Règles Métier Critiques

- Scores calculés UNIQUEMENT côté serveur
- Aucun XP attribué en cas d'échec
- Les bonnes réponses ne sont jamais révélées pendant un quiz actif
- Randomisation des questions ET des réponses côté serveur
- Pas de retour en arrière possible dans un quiz officiel

## Déploiement Vercel

```bash
# 1. Installer Vercel CLI
npm i -g vercel

# 2. Déployer
vercel --prod
```

---

Fait avec ❤️ pour les fans d'anime d'Afrique
