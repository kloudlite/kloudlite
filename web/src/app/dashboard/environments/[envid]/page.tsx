import { redirect } from 'next/navigation';

export default function Page({params}: { params: { envid: string } }) {
  return redirect('/dashboard/environments/' + params.envid + '/apps');
}