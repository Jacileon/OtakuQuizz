# Mini Cup — Tirs au but

Clone du mini-jeu Google Mini Cup (tirs au but) intégré dans l'application existante.

## Stack
- **Frontend** : Next.js 14 + React + Tailwind + shadcn/ui
- **Backend** : Go + Fiber
- **Base de données** : Supabase (PostgreSQL)

## Installation

### 1. Base de données
```bash
# Exécuter la migration
supabase migration up

# Ou manuellement via l'éditeur SQL Supabase :
cat supabase/migrations/040_mini_cup.sql | pbcopy
```

### 2. Backend Go
```bash
# Copier le nouveau handler
cp go/internal/handlers/handlers_minicup.go votre-projet/go/internal/handlers/

# Ajouter les routes dans go/cmd/server/main.go (voir main.go.snippet)

# Relancer le serveur
cd go && go run cmd/server/main.go
```

### 3. Frontend Next.js
```bash
# Copier les fichiers
cp -r src/app/games/mini-cup votre-projet/src/app/games/
cp -r src/components/mini-cup votre-projet/src/components/
cp src/lib/actions/mini-cup.ts votre-projet/src/lib/actions/

# Intégrer le leaderboard dans votre page existante :
# Copier src/app/leaderboard/mini-cup-leaderboard.tsx
# et importer MiniCupLeaderboardTab comme nouvel onglet

# Installer les dépendances si nécessaire
npm install lucide-react  # si pas déjà installé
```

## Fonctionnalités

### Mode Solo vs IA
- 5 tirs réglementaires
- Mort subite en cas d'égalité
- Difficulté progressive du gardien (réaction + précision)
- Dessin/swipe pour viser (souris + tactile)
- Mode clavier accessible (fallback)

### Mode 2 Joueurs Local
- Alternance tireur/gardien
- Écran de transition "Passe l'appareil"
- 5 tirs par équipe puis mort subite

### Système de jeu
- Physique du ballon (gravité, rebond)
- IA gardien avec délai de réaction variable
- Animations : filet qui tremble, plongeon gardien, vague d'impact
- Historique des tirs (pastilles vert/rouge/jaune)
- Score en temps réel

### Intégration
- Gain d'XP par but (+10), victoire (+50), perfect 5/5 (+100)
- Leaderboard dédié avec précision, séries, sessions parfaites
- Historique des parties en base de données
- RLS policies cohérentes avec le reste du projet

## Contrôles

| Périphérique | Action |
|-------------|--------|
| Souris | Pointer down → drag → release pour tirer |
| Tactile | Touch + swipe vers le but |
| Clavier | Activer "Mode clavier" → flèches direction + puissance + TIRER |

## Architecture

```
src/components/mini-cup/
├── MiniCupEngine.tsx      # State machine + canvas + physique
├── PenaltyField.tsx         # (intégré dans Engine) Terrain canvas
├── Ball.tsx                 # (intégré dans Engine) Physique ballon
├── Goalkeeper.tsx           # (intégré dans Engine) IA + animation gardien
├── TeamSelector.tsx         # Sélection des équipes (12 nations)
├── ShotHistory.tsx          # Pastilles de tirs style TV
├── TurnTransitionScreen.tsx # Écran de changement de tour
├── KeyboardAim.tsx          # Fallback clavier accessible
└── GameOverScreen.tsx       # Écran de fin de partie + XP
```

## Assets
Tous les visuels sont générés procéduralement en Canvas 2D (pas de sprites externes). Palette cohérente avec Tailwind/shadcn/ui.

## Tests
1. Aller sur `/games/mini-cup`
2. Choisir un mode puis 2 équipes
3. Dessiner un trait vers le but
4. Vérifier le score, l'historique, la fin de partie
5. Vérifier l'apparition dans `/leaderboard` (onglet Mini Cup)
