"use client";

import { Button } from "@/components/ui/button";
import { toast } from "sonner";

export const ToasterDemo = () => {
  return <Button onClick={()=>{
    toast("Hello world")
  }}>Toaster</Button>
}