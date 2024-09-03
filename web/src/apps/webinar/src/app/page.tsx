import { cn } from 'kl-design-system/utils';
import { JoinWebinar } from './components/join-webinar';


export default function Home() {
  return (
    <main className='flex flex-col h-full'>
      <div className='flex flex-1 flex-col md:items-center self-stretch justify-center px-3xl py-5xl md:py-9xl'>
        <div className='flex flex-col gap-3xl md:w-[500px] px-3xl py-5xl md:px-9xl'>
          <div className='flex flex-col items-stretch"'>
            <div className="flex flex-col gap-lg items-center pb-6xl text-center">
              <div className={cn('text-text-strong headingXl text-center')}>
                Join Kloudlite webinar
              </div>
              <div className="bodyMd-medium text-text-soft">
                Join webinar and experience the power of Kloudlite
              </div>
            </div>
            <JoinWebinar />
          </div>
        </div>
      </div>
    </main >
  );
}
