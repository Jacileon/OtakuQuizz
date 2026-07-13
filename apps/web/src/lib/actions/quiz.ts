'use server';

// ============================================================
// SERVER ACTIONS - CRÉATION DE QUIZ
// ============================================================

import { z } from '../../../node_modules/zod';
import { createClient } from '@/lib/supabase/server';
import { getCurrentUser } from '@/lib/auth/actions';

import { QuizCreateInput, ApiResponse, Quiz } from '@/types';
import { QUIZ_MIN_QUESTIONS } from '@/lib/constants';

const answerSchema = z.object({
  answer_text: z.string().min(1),
  is_correct: z.boolean(),
});

const characterGuessItemSchema = z.object({
  image_url: z.string().optional(),
  answer: z.string().min(1),
  clue: z.string().optional(),
});

const guessClueSchema = z.object({
  type: z.enum(['text', 'image']),
  content: z.string().min(1),
});

const findOddItemSchema = z.object({
  type: z.enum(['text', 'image']),
  content: z.string().min(1),
});

const questionSchema = z.object({
  question_text: z.string().min(1),
  question_type: z.enum(['text', 'true_false', 'image', 'gif', 'audio', 'character_guess', 'impostor']),
  media_url: z.string().optional(),
  media_public_id: z.string().optional(),
  time_limit_seconds: z.number().min(3).max(60).default(30),
  answers: z.array(answerSchema).max(6),
  // character_guess fields
  character_guess_data: z.object({
    characters: z.array(characterGuessItemSchema).min(1).max(4),
  }).optional(),
  character_guess_mode: z.enum(['image', 'text']).optional(),
  // guess (devine) fields
  guess_data: z.object({
    clues: z.array(guessClueSchema).min(1).max(4),
    answer: z.string().min(1),
  }).optional(),
  // find_odd (impostor) fields
  find_odd_data: z.object({
    items: z.array(findOddItemSchema).min(4).max(4),
    odd_index: z.number().min(0).max(3),
  }).optional(),
});

const quizSchema = z.object({
  title: z.string().min(5).max(100),
  description: z.string().max(500).optional(),
  category: z.string().min(1),
  subcategory: z.string().min(1),
  series: z.string().min(1),
  thumbnail_url: z.string().optional(),
  thumbnail_public_id: z.string().optional(),
  questions: z.array(questionSchema).min(QUIZ_MIN_QUESTIONS),
});

export async function createQuiz(formData: QuizCreateInput): Promise<ApiResponse<Quiz>> {
  try {
    const user = await getCurrentUser();
    if (!user) {
      return { data: null, error: 'Non authentifié', success: false };
    }

    const validated = quizSchema.safeParse(formData);
    if (!validated.success) {
      console.error('[QUIZ CREATE] Validation failed:', JSON.stringify(validated.error.errors, null, 2));
      return { data: null, error: `Validation échouée: ${validated.error.errors[0].path.join('.')} - ${validated.error.errors[0].message}`, success: false };
    }

    const supabase = createClient();

    // Vérifier que chaque question a au moins une bonne réponse
    for (const q of validated.data.questions) {
      if (q.question_type === 'character_guess') {
        // Pour character_guess, la réponse est dans character_guess_data
        if (!q.character_guess_data?.characters?.[0]?.answer) {
          return { data: null, error: 'Chaque question character_guess doit avoir une réponse', success: false };
        }
      } else if (q.question_type === 'impostor') {
        // Pour impostor, vérifier find_odd_data
        if (!q.find_odd_data) {
          return { data: null, error: 'Chaque question impostor doit avoir des données find_odd', success: false };
        }
      } else {
        // Pour les autres types, vérifier les answers
        const correctCount = q.answers.filter((a) => a.is_correct).length;
        if (correctCount === 0) {
          return { data: null, error: 'Chaque question doit avoir au moins une bonne réponse', success: false };
        }
      }
    }

    // Insérer le quiz
    const { data: quiz, error: quizError } = await supabase
      .from('quizzes')
      .insert({
        creator_id: user.id,
        title: validated.data.title,
        description: validated.data.description || null,
        category: validated.data.category,
        subcategory: validated.data.subcategory,
        series: validated.data.series,
        thumbnail_url: validated.data.thumbnail_url || null,
        thumbnail_public_id: validated.data.thumbnail_public_id || null,
        question_count: validated.data.questions.length,
        status: 'published',
      })
      .select()
      .single();

    if (quizError || !quiz) {
      console.error('[QUIZ CREATE] Quiz insert error:', quizError?.message, quizError?.details);
      return { data: null, error: `Erreur création quiz: ${quizError?.message || 'Inconnue'}`, success: false };
    }

    // Insérer les questions et réponses
    for (let i = 0; i < validated.data.questions.length; i++) {
      const q = validated.data.questions[i];
      
      // Préparer les données JSON pour les types spéciaux
      let characterGuessData = null;
      let characterGuessMode = null;
      let findOddData = null;

      if (q.question_type === 'character_guess' && q.character_guess_data) {
        characterGuessData = q.character_guess_data;
        characterGuessMode = q.character_guess_mode || 'image';
      }

      if (q.question_type === 'impostor' && q.find_odd_data) {
        findOddData = q.find_odd_data;
      }

      const { data: question, error: qError } = await supabase
        .from('questions')
        .insert({
          quiz_id: quiz.id,
          question_text: q.question_text,
          question_type: q.question_type,
          media_url: q.media_url || null,
          media_public_id: q.media_public_id || null,
          time_limit_seconds: q.time_limit_seconds,
          order_index: i,
          character_guess_data: characterGuessData,
          character_guess_mode: characterGuessMode,
          find_odd_data: findOddData,
        })
        .select()
        .single();

      if (qError || !question) {
        console.error(`[QUIZ CREATE] Question ${i + 1} insert error:`, qError?.message, qError?.details, qError?.hint);
        // Rollback: supprimer le quiz
        await supabase.from('quizzes').delete().eq('id', quiz.id);
        return { data: null, error: `Erreur question ${i + 1}: ${qError?.message || 'Données invalides'}`, success: false };
      }

      // Insérer les réponses selon le type de question
      if (q.question_type === 'character_guess') {
        // Pour character_guess, créer une réponse unique avec le nom du personnage
        const characterAnswer = q.character_guess_data?.characters?.[0]?.answer;
        if (characterAnswer) {
          const { error: aError } = await supabase.from('answers').insert({
            question_id: question.id,
            answer_text: characterAnswer.toUpperCase(),
            is_correct: true,
            order_index: 0,
          });
          if (aError) {
            console.error('[QUIZ CREATE] Answer insert error:', aError?.message, aError?.details, aError?.hint);
            await supabase.from('quizzes').delete().eq('id', quiz.id);
            return { data: null, error: `Erreur réponse: ${aError?.message || 'Inconnue'}`, success: false };
          }
        }
      } else if (q.question_type === 'impostor') {
        // Pour impostor, créer une réponse pour l'intrus
        if (q.find_odd_data) {
          const oddItem = q.find_odd_data.items[q.find_odd_data.odd_index];
          const { error: aError } = await supabase.from('answers').insert({
            question_id: question.id,
            answer_text: oddItem.content,
            is_correct: true,
            order_index: 0,
          });
          if (aError) {
            console.error('[QUIZ CREATE] Answer insert error:', aError?.message, aError?.details, aError?.hint);
            await supabase.from('quizzes').delete().eq('id', quiz.id);
            return { data: null, error: `Erreur réponse: ${aError?.message || 'Inconnue'}`, success: false };
          }
        }
      } else {
        // Insérer les réponses pour les autres types
        const answersToInsert = q.answers.map((a, idx) => ({
          question_id: question.id,
          answer_text: a.answer_text,
          is_correct: a.is_correct,
          order_index: idx,
        }));

        const { error: aError } = await supabase.from('answers').insert(answersToInsert);
        if (aError) {
          console.error('[QUIZ CREATE] Answer insert error:', aError?.message, aError?.details, aError?.hint);
                      await supabase.from('quizzes').delete().eq('id', quiz.id);
                      return { data: null, error: `Erreur réponse: ${aError?.message || 'Inconnue'}`, success: false };
        }
      }
    }

    // Mettre à jour les stats du créateur
    await supabase.rpc('update_user_stats_after_session');

    return { data: quiz as Quiz, error: null, success: true };
  } catch (error) {
    console.error('Erreur createQuiz:', error);
    return { data: null, error: 'Erreur serveur', success: false };
  }
}

export async function updateQuiz(id: string, data: Partial<QuizCreateInput>): Promise<ApiResponse<Quiz>> {
  try {
    const user = await getCurrentUser();
    if (!user) return { data: null, error: 'Non authentifié', success: false };

    const supabase = createClient();

    // Vérifier que l'utilisateur est le créateur
    const { data: existingQuiz } = await supabase
      .from('quizzes')
      .select('id, creator_id')
      .eq('id', id)
      .eq('creator_id', user.id)
      .single();

    if (!existingQuiz) {
      return { data: null, error: 'Quiz non trouvé ou non autorisé', success: false };
    }

    // Mettre à jour les informations du quiz
    const updateData: any = {
      updated_at: new Date().toISOString(),
    };
    if (data.title) updateData.title = data.title;
    if (data.description !== undefined) updateData.description = data.description;
    if (data.category) updateData.category = data.category;
    if (data.subcategory) updateData.subcategory = data.subcategory;
    if (data.series) updateData.series = data.series;
    if (data.thumbnail_url !== undefined) updateData.thumbnail_url = data.thumbnail_url;
    if (data.duration_seconds !== undefined) updateData.duration_seconds = data.duration_seconds;
    if (data.duration_mode !== undefined) updateData.duration_mode = data.duration_mode;

    const { error: updateError } = await supabase
      .from('quizzes')
      .update(updateData)
      .eq('id', id);

    if (updateError) {
      return { data: null, error: 'Erreur mise à jour quiz', success: false };
    }

    // Si des questions sont fournies, les mettre à jour
    if (data.questions && data.questions.length > 0) {
      // Supprimer les anciennes questions et réponses
      const { data: oldQuestions } = await supabase
        .from('questions')
        .select('id')
        .eq('quiz_id', id);

      if (oldQuestions && oldQuestions.length > 0) {
        const oldQuestionIds = oldQuestions.map(q => q.id);
        await supabase.from('answers').delete().in('question_id', oldQuestionIds);
        await supabase.from('questions').delete().eq('quiz_id', id);
      }

      // Insérer les nouvelles questions
      for (let i = 0; i < data.questions.length; i++) {
        const q = data.questions[i];
        const { data: newQuestion, error: qError } = await supabase
          .from('questions')
          .insert({
            quiz_id: id,
            question_text: q.question_text,
            question_type: q.question_type,
            time_limit_seconds: q.time_limit_seconds || 30,
            order_index: i,
            media_url: q.media_url,
            media_public_id: q.media_public_id,
          })
          .select('id')
          .single();

        if (qError) {
          console.error('Erreur création question:', qError);
          continue;
        }

        // Insérer les réponses
        if (q.answers && q.answers.length > 0 && newQuestion) {
          const answersToInsert = q.answers.map((a, j) => ({
            question_id: newQuestion.id,
            answer_text: a.answer_text,
            is_correct: a.is_correct,
            order_index: j,
          }));

          await supabase.from('answers').insert(answersToInsert);
        }
      }

      // Mettre à jour le nombre de questions
      await supabase
        .from('quizzes')
        .update({ question_count: data.questions.length })
        .eq('id', id);
    }

    // Récupérer le quiz mis à jour
    const { data: quiz } = await supabase
      .from('quizzes')
      .select('*')
      .eq('id', id)
      .single();

    return { data: quiz as Quiz, error: null, success: true };
  } catch (error) {
    console.error('Erreur updateQuiz:', error);
    return { data: null, error: 'Erreur serveur', success: false };
  }
}

export async function deleteQuiz(id: string): Promise<ApiResponse<void>> {
  try {
    const user = await getCurrentUser();
    if (!user) return { data: null, error: 'Non authentifié', success: false };

    const supabase = createClient();

    const { error } = await supabase
      .from('quizzes')
      .update({ status: 'deleted', is_visible: false })
      .eq('id', id)
      .eq('creator_id', user.id);

    if (error) {
      return { data: null, error: 'Erreur lors de la suppression', success: false };
    }

    return { data: null, error: null, success: true };
  } catch (error) {
    return { data: null, error: 'Erreur serveur', success: false };
  }
}

export async function archiveQuiz(id: string): Promise<ApiResponse<void>> {
  try {
    const user = await getCurrentUser();
    if (!user) return { data: null, error: 'Non authentifié', success: false };

    const supabase = createClient();

    const { error } = await supabase
      .from('quizzes')
      .update({ status: 'archived', is_visible: false })
      .eq('id', id)
      .eq('creator_id', user.id);

    if (error) {
      return { data: null, error: 'Erreur lors de l\'archivage', success: false };
    }

    return { data: null, error: null, success: true };
  } catch (error) {
    return { data: null, error: 'Erreur serveur', success: false };
  }
}

export async function publishQuiz(id: string): Promise<ApiResponse<Quiz>> {
  try {
    const user = await getCurrentUser();
    if (!user) return { data: null, error: 'Non authentifié', success: false };

    const supabase = createClient();

    const { data: quiz, error } = await supabase
      .from('quizzes')
      .update({ status: 'published', is_visible: true })
      .eq('id', id)
      .eq('creator_id', user.id)
      .select()
      .single();

    if (error || !quiz) {
      return { data: null, error: 'Erreur lors de la publication', success: false };
    }

    return { data: quiz as Quiz, error: null, success: true };
  } catch (error) {
    return { data: null, error: 'Erreur serveur', success: false };
  }
}

export async function deleteQuestion(questionId: string): Promise<void> {
  const user = await getCurrentUser();
  if (!user) throw new Error('Non authentifié');

  const supabase = createClient();

  // Supprimer les réponses associées d'abord
  await supabase
    .from('answers')
    .delete()
    .eq('question_id', questionId);

  // Supprimer la question
  const { error, count } = await supabase
    .from('questions')
    .delete({ count: 'exact' })
    .eq('id', questionId);

  if (error || count === 0) {
    throw new Error('Erreur suppression question');
  }
}
