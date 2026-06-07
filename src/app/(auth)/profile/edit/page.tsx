import { getCurrentProfile } from '@/lib/auth/actions';
import { redirect } from '../../../../../node_modules/next/navigation';
import { ProfileEditForm } from '@/components/profile/ProfileEditForm';
import { User } from 'lucide-react';

export default async function ProfileEditPage() {
  const profile = await getCurrentProfile();
  if (!profile) redirect('/login');

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-2xl mx-auto space-y-6">
        <h1 className="font-display text-3xl tracking-wider flex items-center gap-3">
          <User className="h-8 w-8 text-brand" />
          MODIFIER LE PROFIL
        </h1>
        <ProfileEditForm profile={profile} />
      </div>
    </div>
  );
}