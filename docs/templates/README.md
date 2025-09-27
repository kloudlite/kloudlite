# Code Templates

This directory contains templates for common patterns in the Kloudlite v2 project. Use these templates as starting points when creating new code.

## 📄 Available Templates

### Backend Templates

1. **[backend-service-template.go](./backend-service-template.go)**
   - Complete domain service implementation
   - Clean architecture pattern
   - Repository pattern
   - Business logic validation
   - Dependency injection setup

2. **[grpc-service-template.proto](./grpc-service-template.proto)**
   - gRPC service definition
   - Standard CRUD operations
   - Batch operations
   - Search functionality
   - Proper message structure

### Frontend Templates

3. **[server-action-template.ts](./server-action-template.ts)**
   - Next.js server actions
   - Input validation with Zod
   - Error handling
   - Type-safe responses
   - Cache revalidation

4. **[react-component-template.tsx](./react-component-template.tsx)**
   - Server components
   - Client components
   - Loading states
   - Empty states
   - Search and filters
   - Pagination

## 🚀 How to Use

### Creating a New Backend Service

1. Copy `backend-service-template.go` to your service directory
2. Replace `template` with your service name
3. Update entity fields and business logic
4. Add service-specific validation rules
5. Register in dependency injection

```bash
# Example
cp docs/templates/backend-service-template.go api/apps/myservice/internal/domain/
# Then update package name and customize
```

### Creating a New gRPC Service

1. Copy `grpc-service-template.proto` to `/api/grpc-proto/`
2. Rename and update package/service names
3. Customize messages and methods
4. Generate Go code with protoc
5. Implement the server interface

```bash
# Example
cp docs/templates/grpc-service-template.proto api/grpc-proto/myservice.proto
# Update and generate code
```

### Creating Server Actions

1. Copy `server-action-template.ts` to `/web/app/actions/`
2. Update import paths and types
3. Implement your business logic
4. Add proper validation schemas
5. Handle errors appropriately

```bash
# Example
cp docs/templates/server-action-template.ts web/app/actions/my-feature.ts
```

### Creating React Components

1. Copy relevant sections from `react-component-template.tsx`
2. Choose between server or client component
3. Update types and props
4. Implement component logic
5. Add proper loading/error states

```bash
# Example for a new feature component
cp docs/templates/react-component-template.tsx web/components/my-feature/
```

## 📋 Template Sections

Each template includes:

### Backend Service Template
- Entity definition with validation tags
- Repository interface
- Domain interface and implementation  
- Business validation logic
- Error handling patterns
- Dependency injection module

### gRPC Template
- Service definition with documentation
- Request/Response messages
- Common patterns (CRUD, batch, search)
- Proper enum definitions
- Pagination support
- Error handling guidance

### Server Action Template
- Authentication helpers
- Input validation with Zod
- Type-safe responses
- gRPC client integration
- Cache revalidation
- Comprehensive error handling

### Component Template
- Server component example
- Client component with hooks
- Loading skeletons
- Empty states
- Search functionality
- Filter controls
- Pagination component

## 🔧 Customization Tips

1. **Keep Templates Updated**: As patterns evolve, update templates
2. **Remove Unused Code**: Delete sections you don't need
3. **Follow Conventions**: Ensure naming follows project standards
4. **Add Comments**: Document complex logic
5. **Test Thoroughly**: Templates are starting points, not complete solutions

## 📝 Contributing

When you discover new patterns that should be templated:

1. Create a new template file
2. Add comprehensive comments
3. Include all common variations
4. Update this README
5. Submit a PR with explanation