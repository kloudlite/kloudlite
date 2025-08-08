# Kloudlite v2 Conventions

This directory contains comprehensive conventions and standards for the Kloudlite v2 project. These conventions ensure consistency, maintainability, and quality across the entire codebase.

## 📚 Convention Documents

1. **[Backend Conventions](./backend-conventions.md)** - Go service architecture, patterns, and standards
2. **[Frontend Conventions](./frontend-conventions.md)** - Next.js/React patterns and best practices  
3. **[UI/UX Design System](./design-system.md)** - Visual design standards and component guidelines
4. **[Database Conventions](./database-conventions.md)** - MongoDB schema and query patterns
5. **[API Conventions](./api-conventions.md)** - gRPC, REST, and server action standards
6. **[Git Workflow](./git-workflow.md)** - Branching, commits, and collaboration
7. **[Security Standards](./security-standards.md)** - Authentication, authorization, and security practices

## 🚀 Quick Reference

### Backend Structure
```
/api/apps/{service}/
├── internal/
│   ├── domain/    # Business logic
│   ├── app/       # Use cases
│   └── entities/  # Data models
```

### Frontend Structure  
```
/web/
├── app/          # Pages & routes
├── components/   # UI components
├── lib/          # Utilities
└── hooks/        # Custom hooks
```

### Key Principles

1. **Clean Architecture** - Separate concerns, dependency inversion
2. **Mobile-First Design** - Responsive by default
3. **Type Safety** - TypeScript/Go types everywhere
4. **Server-First** - SSR/SSG over client rendering
5. **Security-First** - Auth checks at boundaries

## 🔧 Getting Started

1. Read the relevant convention documents for your area
2. Use the provided templates in `/templates`
3. Follow the linting rules configured in the project
4. Ask questions if conventions are unclear

## 📝 Contributing

When proposing changes to conventions:
1. Create a PR with clear rationale
2. Update affected documentation
3. Get team consensus before merging