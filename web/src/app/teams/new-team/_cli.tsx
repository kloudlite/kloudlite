"use client";

import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { Form, FormField } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { PopoverContent } from "@/components/ui/popover";
import { cn } from "@/lib/utils";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Popover, PopoverTrigger } from "@radix-ui/react-popover";
import { SelectGroup } from "@radix-ui/react-select";
import { Check, ChevronsUpDown } from "lucide-react";
import Link from "next/link";
import { useCallback, useState } from "react";
import { useForm } from "react-hook-form";

const regions = [
  { label: "US East (Ohio)", value: "us-east-2" },
  { label: "US East (N. Virginia)", value: "us-east-1" },
  { label: "US West (N. California)", value: "us-west-1" },
  { label: "US West (Oregon)", value: "us-west-2" },
  { label: "Canada (Central)", value: "ca-central-1" },
  { label: "Canada West (Calgary)", value: "ca-west-1" },
  { label: "Mexico (Central)", value: "mx-central-1" },
  { label: "South America (São Paulo)", value: "sa-east-1" },
  { label: "Europe (Frankfurt)", value: "eu-central-1" },
  { label: "Europe (Ireland)", value: "eu-west-1" },
  { label: "Europe (London)", value: "eu-west-2" },
  { label: "Europe (Milan)", value: "eu-south-1" },
  { label: "Europe (Paris)", value: "eu-west-3" },
  { label: "Europe (Spain)", value: "eu-south-2" },
  { label: "Europe (Stockholm)", value: "eu-north-1" },
  { label: "Europe (Zurich)", value: "eu-central-2" },
  { label: "Asia Pacific (Hong Kong)", value: "ap-east-1" },
  { label: "Asia Pacific (Hyderabad)", value: "ap-south-2" },
  { label: "Asia Pacific (Jakarta)", value: "ap-southeast-3" },
  { label: "Asia Pacific (Malaysia)", value: "ap-southeast-5" },
  { label: "Asia Pacific (Melbourne)", value: "ap-southeast-4" },
  { label: "Asia Pacific (Mumbai)", value: "ap-south-1" },
  { label: "Asia Pacific (Osaka)", value: "ap-northeast-3" },
  { label: "Asia Pacific (Seoul)", value: "ap-northeast-2" },
  { label: "Asia Pacific (Singapore)", value: "ap-southeast-1" },
  { label: "Asia Pacific (Sydney)", value: "ap-southeast-2" },
  { label: "Asia Pacific (Thailand)", value: "ap-southeast-7" },
  { label: "Asia Pacific (Tokyo)", value: "ap-northeast-1" },
  { label: "Middle East (Bahrain)", value: "me-south-1" },
  { label: "Middle East (UAE)", value: "me-central-1" },
  { label: "Israel (Tel Aviv)", value: "il-central-1" },
  { label: "Africa (Cape Town)", value: "af-south-1" },
  { label: "AWS GovCloud (US-East)", value: "us-gov-east-1" },
  { label: "AWS GovCloud (US-West)", value: "us-gov-west-1" },
];

const RegionSelection = (
  { value, setValue }: { value: string; setValue: (k: string) => void },
) => {
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="w-full justify-between"
        >
          {value
            ? regions.find((region) => {
              return region.value === value
            })?.label
            : "Select region..."}
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-full p-0" align="start">
        <Command>
          <CommandInput placeholder="Search regions..." />
          <CommandList>
            <CommandEmpty>No region found.</CommandEmpty>
            <CommandGroup>
              {regions.map((region) => (
                <CommandItem
                  key={region.value}
                  value={region.value}
                  onSelect={(currentValue) => {
                    setValue(currentValue);
                    setOpen(false);
                  }}
                  keywords={[region.label.toLowerCase()]}
                >
                  <Check
                    className={cn(
                      "mr-2 h-4 w-4",
                      value === region.value ? "opacity-100" : "opacity-0",
                    )}
                  />
                  <span>
                    {region.label}
                  </span>
                  <span className="text-muted-foreground">
                    {region.value}
                  </span>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
};

export const TeamCreationForm = () => {
  const form = useForm({
    defaultValues: {
      teamName: "",
      region: "",
    }
  });
  return (
    <Form {...form}>
      <div className="flex flex-col gap-4">
        <FormField
          name="teamName"
          render={() => {
            return (
              <div className="flex flex-col gap-2">
                <label htmlFor="teamName" className="text-sm font-medium">
                  Team Name
                </label>
                <Input
                  id="teamName"
                  placeholder="Team Name"
                  className="border p-2 rounded-md w-full"
                  {...form.register("teamName", {
                    required: "Team name is required",
                  })}
                  type="text"
                  autoComplete="off"
                  autoCorrect="off"
                  autoCapitalize="none"
                  spellCheck="false"
                  maxLength={50}
                  minLength={3}
                  pattern="^[a-zA-Z0-9_ ]{3,50}$"
                />
              </div>
            );
          }}
        />
        <FormField
          name="region"
          render={() => {
            const regionForm = useCallback(() => {
              form.register("region", {
                required: "Choose Region",
              });
            }, []);
            return (
              <div className="flex flex-col gap-2">
                <label htmlFor="region" className="text-sm font-medium">
                  Region
                </label>
                <RegionSelection
                  value={form.watch("region")}
                  setValue={(value) => {
                    console.log("Selected region:", value);
                    form.setValue("region", value);
                  }}
                />
              </div>
            );
          }}
        />
        <div className="flex justify-end">
          <Button
            variant="outline"
            className="mr-2"
            onClick={() => {
              form.reset();
            }}
          >
            Cancel
          </Button>
          <Button asChild>
            <Link href={"/dashboard"}>
              Create Team
            </Link>
          </Button>
        </div>
      </div>
    </Form>
  );
};
