import { Navbar } from '@/components/home/navbar'
import { HeroSection } from '@/components/home/hero-section'
import { Footer } from '@/components/home/footer'

export default function Home() {
  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main>
        <HeroSection />
      </main>
      <Footer />
    </div>
  );
}