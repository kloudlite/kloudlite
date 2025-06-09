import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { TeamCreationForm } from "./_cli";

export default function NewTeam() {
  
  return (
    <div className="w-[700px] p-12 flex flex-col gap-4">
      <Card className="mt-20">
        <CardHeader>
          <CardTitle>
            Setup your team!
          </CardTitle>
          <CardDescription>Configure team to start</CardDescription>
        </CardHeader>
        <CardContent>
          <TeamCreationForm />
        </CardContent>
      </Card>
    </div>
  );
}
