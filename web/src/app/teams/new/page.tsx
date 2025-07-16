import { Button } from '@/components/ui/button'
import { Link } from '@/components/ui/link'
import { ArrowLeft, Users } from 'lucide-react'
import { CreateTeamForm } from '@/components/teams/create-team-form'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export default function NewTeamPage() {
  return (
    <div className="min-h-screen flex items-center justify-center p-4 relative">
      <div className="absolute top-4 left-4">
        <Button variant="ghost" size="sm" asChild>
          <Link href="/teams">
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Teams
          </Link>
        </Button>
      </div>
      
      <div className="w-full max-w-md">
        <Card className="w-full">
          <CardHeader className="space-y-4 pb-8 text-center">
            <div className="mx-auto w-12 h-12 flex items-center justify-center">
              <Users className="h-12 w-12 text-primary" strokeWidth={1.5} />
            </div>
            <div className="space-y-2">
              <CardTitle className="text-3xl font-semibold tracking-tight">
                Create New Team
              </CardTitle>
              <CardDescription className="text-base text-muted-foreground">
                Set up a new team to collaborate with others
              </CardDescription>
            </div>
          </CardHeader>
          <CardContent className="pb-8">
            <CreateTeamForm />
          </CardContent>
        </Card>
      </div>
    </div>
  )
}