'use client';

import { useState } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { ChevronDown, ChevronRight, HelpCircle } from 'lucide-react';
import { cn } from '@/lib/utils';

const faqData = [
  {
    theme: 'Concept du site',
    questions: [
      {
        question: "Qu'est-ce que Otaku Quiz Africa ?",
        answer: "Otaku Quiz Africa est une plateforme de quiz dédiée à la culture anime/manga en Afrique. Elle permet de tester ses connaissances, de défier ses amis, de créer ses propres quiz et de progresser dans un système de rangs."
      },
      {
        question: "À qui s'adresse la plateforme ?",
        answer: "La plateforme s'adresse à tous les fans d'anime et manga en Afrique et dans le monde francophone. Que vous soyez débutant ou expert, il y a des quiz pour tous les niveaux."
      },
      {
        question: "Comment fonctionne la progression ?",
        answer: "Vous gagnez de l'XP en jouant aux quiz. Plus vous avez de bonnes réponses et plus vous répondez vite, plus vous gagnez d'XP. L'XP vous permet de monter en rang et de débloquer de nouvelles fonctionnalités."
      }
    ]
  },
  {
    theme: 'Ce que vous pouvez faire',
    questions: [
      {
        question: "Puis-je jouer à des quiz seul ?",
        answer: "Oui ! Vous pouvez jouer à tous les quiz en mode solo. Les quiz de type 'Défi' nécessitent au minimum 2 joueurs."
      },
      {
        question: "Comment défier mes amis ?",
        answer: "Sur la page d'un quiz, cliquez sur 'Défier vos amis'. Sélectionnez les amis à inviter, définissez votre mise XP et envoyez les invitations. Vos amis pourront accepter, refuser ou modifier leur mise."
      },
      {
        question: "Puis-je créer mes propres quiz ?",
        answer: "La création de quiz est réservée aux utilisateurs ayant le rang C, E ou S par défaut. Les admins peuvent aussi accorder cette permission individuellement. Si vous n'avez pas le rang requis, vous verrez un message vous indiquant le rang à atteindre."
      },
      {
        question: "Comment gagner de l'XP ?",
        answer: "Vous gagnez de l'XP en complétant des quiz. Le montant d'XP gagné dépend de votre score, de votre précision et de votre rapidité. Les quiz officiels et les défis peuvent rapporter plus d'XP."
      }
    ]
  },
  {
    theme: 'Ce qui est limité ou interdit',
    questions: [
      {
        question: "Y a-t-il une limite de participations aux défis ?",
        answer: "Oui, vous ne pouvez participer à un défi sur le même quiz que 3 fois maximum. Cette limite s'applique à tous les participants du défi."
      },
      {
        question: "Puis-je miser plus d'XP que mon solde ?",
        answer: "Non, il est impossible de miser plus d'XP que votre solde disponible. La vérification est faite côté serveur pour garantir l'équité."
      },
      {
        question: "Puis-je créer un quiz sans le rang requis ?",
        answer: "Non, sauf si un admin vous a accordé une autorisation individuelle. Sinon, vous devez atteindre le rang requis (C, E ou S par défaut)."
      }
    ]
  },
  {
    theme: 'Comment se déroule un quiz',
    questions: [
      {
        question: "Comment fonctionne le timer ?",
        answer: "Chaque question a un temps limité. Un décompte visible s'affiche pendant la session. Si le temps expire, la soumission est automatique et la question est comptée comme non répondue."
      },
      {
        question: "Quels types de questions sont disponibles ?",
        answer: "Il existe plusieurs types : QCM classique, Vrai/Faux, Image, GIF, Audio, 'Devine le personnage' et 'Trouve l'intrus'. Chaque type a ses propres règles d'affichage."
      },
      {
        question: "Comment sont calculés les points ?",
        answer: "Les points sont calculés en fonction de la correction de la réponse, de la rapidité et d'un bonus de série (streak). Plus vous répondez correctement et vite, plus vous gagnez de points."
      }
    ]
  },
  {
    theme: 'Événements officiels',
    questions: [
      {
        question: "Qu'est-ce qu'un quiz officiel ?",
        answer: "Les quiz officiels sont créés uniquement par les admins. Ils sont identifiables par le badge 'Officiel' et peuvent avoir des récompenses spéciales. Leur classement est public et permanent."
      },
      {
        question: "Comment fonctionnent les récompenses ?",
        answer: "Les admins peuvent configurer des récompenses pour les quiz officiels. Les 3 premières places sont affichées sur un podium visuel (or/argent/bronze). Les récompenses sont attribuées automatiquement à la clôture du quiz."
      },
      {
        question: "Les classements officiels sont-ils permanents ?",
        answer: "Oui, le classement des quiz officiels est toujours public et accessible, même après l'archivage du quiz."
      }
    ]
  },
  {
    theme: 'Système de rangs',
    questions: [
      {
        question: "Quels sont les rangs disponibles ?",
        answer: "Les rangs sont : F, E, D, C, B, A, S, S+, SS, SSS et Légende. Chaque rang a un seuil d'XP minimum à atteindre."
      },
      {
        question: "Comment progresser en rang ?",
        answer: "Vous progressez en gagnant de l'XP. Plus vous jouez et plus vos scores sont élevés, plus vous gagnez d'XP et montez en rang."
      },
      {
        question: "Que débloque chaque rang ?",
        answer: "Certains rangs débloquent des fonctionnalités. Par exemple, le rang C permet de créer des quiz. Le rang Légende donne accès à des privilèges spéciaux."
      }
    ]
  }
];

export default function FaqPage() {
  const [openThemes, setOpenThemes] = useState<Set<string>>(new Set());
  const [openQuestions, setOpenQuestions] = useState<Set<string>>(new Set());

  const toggleTheme = (theme: string) => {
    setOpenThemes(prev => {
      const newSet = new Set(prev);
      if (newSet.has(theme)) {
        newSet.delete(theme);
      } else {
        newSet.add(theme);
      }
      return newSet;
    });
  };

  const toggleQuestion = (question: string) => {
    setOpenQuestions(prev => {
      const newSet = new Set(prev);
      if (newSet.has(question)) {
        newSet.delete(question);
      } else {
        newSet.add(question);
      }
      return newSet;
    });
  };

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-3xl mx-auto">
        <div className="mb-8">
          <h1 className="font-display text-3xl tracking-wider flex items-center gap-3">
            <HelpCircle className="h-8 w-8 text-primary" />
            FAQ
          </h1>
          <p className="text-muted-foreground mt-2">
            Trouvez les réponses à vos questions
          </p>
        </div>

        <div className="space-y-4">
          {faqData.map((section) => (
            <Card key={section.theme}>
              <CardContent className="p-0">
                <button
                  onClick={() => toggleTheme(section.theme)}
                  className="w-full flex items-center justify-between p-4 hover:bg-accent/50 transition-colors"
                >
                  <h2 className="font-semibold text-left">{section.theme}</h2>
                  {openThemes.has(section.theme) ? (
                    <ChevronDown className="h-5 w-5 shrink-0" />
                  ) : (
                    <ChevronRight className="h-5 w-5 shrink-0" />
                  )}
                </button>

                {openThemes.has(section.theme) && (
                  <div className="border-t">
                    {section.questions.map((qa) => (
                      <div key={qa.question} className="border-b last:border-b-0">
                        <button
                          onClick={() => toggleQuestion(qa.question)}
                          className="w-full flex items-center justify-between p-4 pl-8 hover:bg-accent/30 transition-colors"
                        >
                          <span className="text-sm font-medium text-left">{qa.question}</span>
                          {openQuestions.has(qa.question) ? (
                            <ChevronDown className="h-4 w-4 shrink-0" />
                          ) : (
                            <ChevronRight className="h-4 w-4 shrink-0" />
                          )}
                        </button>

                        {openQuestions.has(qa.question) && (
                          <div className="px-8 pb-4">
                            <p className="text-sm text-muted-foreground">{qa.answer}</p>
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </div>
  );
}