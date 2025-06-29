"use client";

import { useState } from "react";
import { Label, Input, Button } from "@/components/atoms";
import { TextField, SelectField, IconInput, AuthFormWrapper } from "@/components/molecules";
import { Mail, Lock, User, Search } from "lucide-react";
import { ComponentShowcase } from "../../_components/component-showcase";

export default function FormsPage() {
  const [selectValue, setSelectValue] = useState("");

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">
          Form Components
        </h1>
        <p className="text-slate-600 dark:text-slate-400">
          Enhanced form components for better user experience.
        </p>
      </div>

      <ComponentShowcase
        title="Text Fields"
        description="Input fields with integrated labels and error states"
      >
        <div className="space-y-4 max-w-md">
          <TextField
            label="Username"
            placeholder="Enter your username"
            helperText="Choose a unique username"
          />
          
          <TextField
            label="Email"
            type="email"
            placeholder="you@example.com"
            required
          />
          
          <TextField
            label="Password"
            type="password"
            placeholder="••••••••"
            error="Password must be at least 8 characters"
          />
          
          <TextField
            label="Disabled Field"
            placeholder="Cannot edit this"
            disabled
          />
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Select Fields"
        description="Dropdown select components"
      >
        <div className="space-y-4 max-w-md">
          <SelectField
            label="Choose Option"
            value={selectValue}
            onValueChange={setSelectValue}
            placeholder="Select an option"
            options={[
              { value: "option1", label: "Option 1" },
              { value: "option2", label: "Option 2" },
              { value: "option3", label: "Option 3" },
            ]}
          />
          
          <SelectField
            label="Required Select"
            value=""
            onValueChange={() => {}}
            placeholder="Select a value"
            required
            options={[
              { value: "value1", label: "Value 1" },
              { value: "value2", label: "Value 2" },
            ]}
          />
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Icon Inputs"
        description="Input fields with leading icons"
      >
        <div className="space-y-4 max-w-md">
          <div>
            <Label>Search</Label>
            <IconInput
              icon={Search}
              placeholder="Search items..."
              className="mt-1"
            />
          </div>
          
          <div>
            <Label>Email</Label>
            <IconInput
              icon={Mail}
              type="email"
              placeholder="you@example.com"
              className="mt-1"
            />
          </div>
          
          <div>
            <Label>Password</Label>
            <IconInput
              icon={Lock}
              type="password"
              placeholder="••••••••"
              className="mt-1"
            />
          </div>
          
          <div>
            <Label>Username</Label>
            <IconInput
              icon={User}
              placeholder="johndoe"
              className="mt-1"
            />
          </div>
        </div>
      </ComponentShowcase>

      <ComponentShowcase
        title="Auth Form Wrapper"
        description="Pre-styled form wrapper for authentication pages"
      >
        <div className="max-w-md mx-auto">
          <AuthFormWrapper
            title="Sign In"
            subtitle="Enter your credentials to access your account"
            footerText="Don't have an account?"
            footerLinkText="Sign up"
            footerLinkHref="/signup"
          >
            <TextField
              label="Email"
              type="email"
              placeholder="you@example.com"
            />
            <TextField
              label="Password"
              type="password"
              placeholder="••••••••"
            />
            <Button className="w-full">Sign In</Button>
          </AuthFormWrapper>
        </div>
      </ComponentShowcase>
    </div>
  );
}