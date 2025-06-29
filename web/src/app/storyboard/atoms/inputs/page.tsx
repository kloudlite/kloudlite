"use client";

import { 
  Input, 
  Label, 
  Caption, 
  Textarea, 
  Checkbox, 
  RadioGroup, 
  RadioGroupItem, 
  Switch,
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
  MultiSelect,
  type MultiSelectOption
} from "@/components/atoms";
import { Mail, Search, Lock, User } from "lucide-react";
import { ComponentShowcase } from "../../_components/component-showcase";
import { useState } from "react";

export default function InputsPage() {
  const [switchValue, setSwitchValue] = useState(false);
  const [checkboxValue, setCheckboxValue] = useState(false);
  const [radioValue, setRadioValue] = useState("option1");
  const [selectValue, setSelectValue] = useState("");
  const [multiSelectValues, setMultiSelectValues] = useState<string[]>([]);

  const multiSelectOptions: MultiSelectOption[] = [
    { label: "React", value: "react" },
    { label: "Vue", value: "vue" },
    { label: "Angular", value: "angular" },
    { label: "Svelte", value: "svelte" },
    { label: "Next.js", value: "nextjs" },
    { label: "Nuxt", value: "nuxt" },
    { label: "Gatsby", value: "gatsby" },
    { label: "Remix", value: "remix" },
    { label: "TypeScript", value: "typescript" },
    { label: "JavaScript", value: "javascript" },
  ];

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-foreground mb-4">
          Input Components
        </h1>
        <p className="text-muted-foreground">
          Form input components for user data entry.
        </p>
      </div>

      <ComponentShowcase
        title="Text Inputs"
        description="Basic text input variations"
      >
        <div className="space-y-4 max-w-md">
          <div>
            <Label htmlFor="default">Default Input</Label>
            <Input id="default" placeholder="Enter text..." className="mt-2" />
          </div>
          
          <div>
            <Label htmlFor="with-icon">With Icon</Label>
            <div className="relative mt-2 group">
              <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground group-focus-within:text-primary transition-colors duration-200" />
              <Input id="with-icon" placeholder="Email address" className="pl-10" />
            </div>
          </div>

          <div>
            <Label htmlFor="disabled">Disabled</Label>
            <Input id="disabled" placeholder="Disabled input" disabled className="mt-2" />
          </div>

          <div>
            <Label htmlFor="with-error">With Error</Label>
            <Input id="with-error" placeholder="Invalid input" error className="mt-2" />
            <Caption error className="mt-2">This field is required</Caption>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Input Types"
        description="Different HTML input types"
      >
        <div className="space-y-4 max-w-md">
          <div>
            <Label htmlFor="email">Email</Label>
            <Input id="email" type="email" placeholder="you@example.com" className="mt-2" />
          </div>
          
          <div>
            <Label htmlFor="password">Password</Label>
            <div className="relative mt-2 group">
              <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground group-focus-within:text-primary transition-colors duration-200" />
              <Input id="password" type="password" placeholder="••••••••" className="pl-10" />
            </div>
          </div>

          <div>
            <Label htmlFor="number">Number</Label>
            <Input id="number" type="number" placeholder="0" className="mt-2" />
          </div>

          <div>
            <Label htmlFor="date">Date</Label>
            <Input id="date" type="date" className="mt-2" />
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Textarea"
        description="Multi-line text input"
      >
        <div className="max-w-md">
          <Label htmlFor="textarea">Description</Label>
          <Textarea 
            id="textarea" 
            placeholder="Enter a detailed description..." 
            className="mt-2"
            rows={4}
          />
          <Caption className="mt-2">Maximum 500 characters</Caption>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Checkboxes"
        description="Single and multiple selection"
      >
        <div className="space-y-4">
          <div className="flex items-center space-x-2">
            <Checkbox 
              id="terms" 
              checked={checkboxValue}
              onCheckedChange={(checked) => setCheckboxValue(checked as boolean)}
            />
            <Label htmlFor="terms" className="text-sm font-normal">
              Accept terms and conditions
            </Label>
          </div>

          <div className="space-y-2">
            <Label>Select options</Label>
            <div className="space-y-2">
              <div className="flex items-center space-x-2">
                <Checkbox id="option1" />
                <Label htmlFor="option1" className="text-sm font-normal">Option 1</Label>
              </div>
              <div className="flex items-center space-x-2">
                <Checkbox id="option2" />
                <Label htmlFor="option2" className="text-sm font-normal">Option 2</Label>
              </div>
              <div className="flex items-center space-x-2">
                <Checkbox id="option3" disabled />
                <Label htmlFor="option3" className="text-sm font-normal text-slate-400">
                  Option 3 (disabled)
                </Label>
              </div>
            </div>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Radio Groups"
        description="Single selection from multiple options"
      >
        <div>
          <Label>Select size</Label>
          <RadioGroup value={radioValue} onValueChange={setRadioValue} className="mt-2">
            <div className="flex items-center space-x-2">
              <RadioGroupItem value="small" id="small" />
              <Label htmlFor="small" className="font-normal">Small</Label>
            </div>
            <div className="flex items-center space-x-2">
              <RadioGroupItem value="medium" id="medium" />
              <Label htmlFor="medium" className="font-normal">Medium</Label>
            </div>
            <div className="flex items-center space-x-2">
              <RadioGroupItem value="large" id="large" />
              <Label htmlFor="large" className="font-normal">Large</Label>
            </div>
          </RadioGroup>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Switches"
        description="Toggle switches for on/off states"
      >
        <div className="space-y-4">
          <div className="flex items-center space-x-2">
            <Switch 
              id="airplane-mode"
              checked={switchValue}
              onCheckedChange={setSwitchValue}
            />
            <Label htmlFor="airplane-mode">Airplane Mode</Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch id="notifications" defaultChecked />
            <Label htmlFor="notifications">Enable notifications</Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch id="disabled-switch" disabled />
            <Label htmlFor="disabled-switch" className="text-slate-400">
              Disabled switch
            </Label>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Select"
        description="Single selection dropdown"
      >
        <div className="space-y-4 max-w-md">
          <div>
            <Label htmlFor="select-default">Choose a framework</Label>
            <Select value={selectValue} onValueChange={setSelectValue}>
              <SelectTrigger id="select-default" className="mt-2">
                <SelectValue placeholder="Select a framework" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectLabel>Frontend Frameworks</SelectLabel>
                  <SelectItem value="react">React</SelectItem>
                  <SelectItem value="vue">Vue</SelectItem>
                  <SelectItem value="angular">Angular</SelectItem>
                  <SelectItem value="svelte">Svelte</SelectItem>
                </SelectGroup>
                <SelectGroup>
                  <SelectLabel>Meta Frameworks</SelectLabel>
                  <SelectItem value="nextjs">Next.js</SelectItem>
                  <SelectItem value="nuxt">Nuxt</SelectItem>
                  <SelectItem value="gatsby">Gatsby</SelectItem>
                  <SelectItem value="remix">Remix</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>

          <div>
            <Label htmlFor="select-disabled">Disabled Select</Label>
            <Select disabled>
              <SelectTrigger id="select-disabled" className="mt-2">
                <SelectValue placeholder="Select is disabled" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="option1">Option 1</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div>
            <Label htmlFor="select-error">With Error</Label>
            <Select>
              <SelectTrigger id="select-error" className="mt-2 border-destructive" aria-invalid="true">
                <SelectValue placeholder="Invalid selection" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="option1">Option 1</SelectItem>
                <SelectItem value="option2">Option 2</SelectItem>
              </SelectContent>
            </Select>
            <Caption error className="mt-2">Please select a valid option</Caption>
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="MultiSelect"
        description="Multiple selection dropdown with search"
      >
        <div className="space-y-4 max-w-md">
          <div>
            <Label htmlFor="multiselect-default">Select technologies</Label>
            <MultiSelect
              options={multiSelectOptions}
              selected={multiSelectValues}
              onChange={setMultiSelectValues}
              placeholder="Select technologies..."
              className="mt-2"
            />
            <Caption className="mt-2">You can select multiple options</Caption>
          </div>

          <div>
            <Label htmlFor="multiselect-with-search">With Search Enabled</Label>
            <MultiSelect
              options={multiSelectOptions}
              selected={multiSelectValues}
              onChange={setMultiSelectValues}
              placeholder="Select technologies..."
              enableSearch={true}
              className="mt-2"
            />
            <Caption className="mt-2">Search through options</Caption>
          </div>

          <div>
            <Label htmlFor="multiselect-preselected">With Pre-selected Values</Label>
            <MultiSelect
              options={multiSelectOptions}
              selected={["react", "nextjs", "typescript"]}
              onChange={() => {}}
              placeholder="Select technologies..."
              className="mt-2"
            />
          </div>

          <div>
            <Label htmlFor="multiselect-disabled">Disabled MultiSelect</Label>
            <MultiSelect
              options={multiSelectOptions}
              selected={[]}
              onChange={() => {}}
              placeholder="MultiSelect is disabled..."
              disabled
              className="mt-2"
            />
          </div>

          <div>
            <Label htmlFor="multiselect-error">With Error</Label>
            <MultiSelect
              options={multiSelectOptions}
              selected={[]}
              onChange={() => {}}
              placeholder="Invalid selection..."
              className="mt-2 border-destructive"
              aria-invalid="true"
            />
            <Caption error className="mt-2">At least one technology must be selected</Caption>
          </div>
        </div>
      </ComponentShowcase>
    </div>
  );
}