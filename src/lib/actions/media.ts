'use server';

// ============================================================
// SERVER ACTIONS - UPLOAD MÉDIA
// ============================================================

import { v2 as cloudinary } from 'cloudinary';

cloudinary.config({
  cloud_name: process.env.NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME,
  api_key: process.env.CLOUDINARY_API_KEY,
  api_secret: process.env.CLOUDINARY_API_SECRET,
});

export async function uploadQuizMedia(
  file: File,
  type: 'image' | 'audio' | 'gif'
): Promise<{ url: string; publicId: string } | { error: string }> {
  try {
    const bytes = await file.arrayBuffer();
    const buffer = Buffer.from(bytes);
    const base64 = buffer.toString('base64');
    const dataURI = `data:${file.type};base64,${base64}`;

    let result;

    if (type === 'image' || type === 'gif') {
      result = await cloudinary.uploader.upload(dataURI, {
        folder: 'otaku-quiz/quiz-media',
        transformation: [
          { width: 800, height: 600, crop: 'limit' },
          { quality: 'auto', fetch_format: 'auto' },
        ],
      });
    } else if (type === 'audio') {
      // Vérifier la durée côté serveur
      result = await cloudinary.uploader.upload(dataURI, {
        folder: 'otaku-quiz/quiz-audio',
        resource_type: 'video', // Cloudinary gère l'audio comme video
      });
    }

    if (!result) {
      return { error: 'Upload failed' };
    }

    return {
      url: result.secure_url,
      publicId: result.public_id,
    };
  } catch (error) {
    return { error: 'Erreur upload' };
  }
}

export async function deleteMedia(publicId: string): Promise<void> {
  try {
    await cloudinary.uploader.destroy(publicId);
  } catch (error) {
    console.error('Erreur suppression média:', error);
  }
}

export async function uploadAvatar(
  file: File
): Promise<{ url: string; publicId: string } | { error: string }> {
  try {
    const bytes = await file.arrayBuffer();
    const buffer = Buffer.from(bytes);
    const base64 = buffer.toString('base64');
    const dataURI = `data:${file.type};base64,${base64}`;

    const result = await cloudinary.uploader.upload(dataURI, {
      folder: 'otaku-quiz/avatars',
      transformation: [
        { width: 200, height: 200, crop: 'fill', gravity: 'face' },
        { quality: 'auto', fetch_format: 'auto' },
      ],
    });

    return {
      url: result.secure_url,
      publicId: result.public_id,
    };
  } catch (error) {
    console.error('Erreur upload avatar:', error);
    return { error: 'Erreur upload avatar' };
  }
}

