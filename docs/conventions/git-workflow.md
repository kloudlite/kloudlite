# Git Workflow

This document outlines the Git workflow, branching strategy, and collaboration conventions for Kloudlite v2.

## 🌳 Branching Strategy

### Main Branches
- **`master`** - Production-ready code
- **`develop`** - Integration branch for features (optional)

### Feature Branches
```bash
# Feature development
feat/team-management
feat/notification-system
feat/oauth-integration

# Bug fixes
fix/login-error
fix/team-slug-validation
fix/notification-duplicate

# Improvements
chore/update-dependencies
chore/cleanup-imports
refactor/auth-service
docs/api-documentation
```

### Branch Naming Convention
- **Format**: `{type}/{description}`
- **Types**:
  - `feat` - New features
  - `fix` - Bug fixes
  - `refactor` - Code refactoring
  - `chore` - Maintenance tasks
  - `docs` - Documentation only
  - `test` - Test additions/changes
  - `perf` - Performance improvements

## 📝 Commit Conventions

### Commit Message Format
```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types
- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, etc)
- **refactor**: Code refactoring
- **perf**: Performance improvements
- **test**: Test changes
- **chore**: Maintenance tasks

### Examples
```bash
# Simple commits
feat: implement team approval workflow
fix: resolve duplicate notification issue
docs: update API documentation

# With scope
feat(auth): add OAuth provider configuration
fix(teams): correct slug validation regex
refactor(frontend): optimize bundle size

# With body
feat(notifications): implement role-based targeting

- Add support for team_role and platform_role targeting
- Include deduplication mechanism
- Auto-mark notifications as read when viewed

# With breaking change
feat(api)!: change team API response format

BREAKING CHANGE: Team API now returns nested user object
instead of just userId
```

### Commit Guidelines
1. **Keep commits atomic** - One logical change per commit
2. **Write clear messages** - Explain what and why
3. **Use present tense** - "add" not "added"
4. **Keep subject line under 50 chars**
5. **Add body for complex changes**
6. **Reference issues** - "Fixes #123"

## 🔄 Pull Request Process

### PR Title Format
Same as commit message format:
```
feat(teams): implement team management system
fix(auth): resolve session timeout issue
```

### PR Template
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change)
- [ ] New feature (non-breaking change)
- [ ] Breaking change
- [ ] Documentation update

## Changes Made
- List specific changes
- Include technical details
- Mention any dependencies

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Screenshots (if applicable)
Add screenshots for UI changes

## Checklist
- [ ] Code follows project conventions
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] No console.logs or debug code
- [ ] Linting passes
```

### PR Guidelines
1. **Keep PRs focused** - One feature/fix per PR
2. **Write descriptive titles**
3. **Add comprehensive description**
4. **Include test plan**
5. **Request appropriate reviewers**
6. **Respond to feedback promptly**
7. **Squash commits when merging**

## 👀 Code Review

### Review Checklist
- [ ] Code follows conventions
- [ ] Logic is correct
- [ ] Edge cases handled
- [ ] Error handling appropriate
- [ ] Performance considerations
- [ ] Security implications
- [ ] Tests included/updated
- [ ] Documentation updated

### Review Comments
```go
// Suggest improvements
// suggestion: Consider using a map for O(1) lookup instead of slice iteration

// Ask questions
// question: Why do we need this additional check here?

// Highlight issues
// issue: This could cause a race condition when accessed concurrently

// Praise good code
// praise: Nice use of the factory pattern here!
```

## 🔧 Git Configuration

### Recommended Settings
```bash
# Set your identity
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Helpful aliases
git config --global alias.co checkout
git config --global alias.br branch
git config --global alias.ci commit
git config --global alias.st status
git config --global alias.last 'log -1 HEAD'
git config --global alias.unstage 'reset HEAD --'

# Better diffs
git config --global diff.algorithm histogram

# Rebase by default
git config --global pull.rebase true
```

### .gitignore Standards
```gitignore
# Dependencies
node_modules/
vendor/

# Build outputs
dist/
build/
bin/
*.exe

# Environment files
.env
.env.local
.env.*.local

# IDE files
.idea/
.vscode/
*.swp
*.swo
*~

# OS files
.DS_Store
Thumbs.db

# Logs
*.log
npm-debug.log*
yarn-error.log*

# Test coverage
coverage/
*.cover
.nyc_output/

# Temporary files
tmp/
temp/
```

## 🚀 Workflow Examples

### Feature Development
```bash
# 1. Create feature branch
git checkout -b feat/team-management

# 2. Make changes and commit
git add .
git commit -m "feat(teams): add team creation API"

# 3. Push branch
git push -u origin feat/team-management

# 4. Create PR
# Use GitHub/GitLab UI or CLI

# 5. After approval, merge
# Preferably squash and merge
```

### Hotfix Process
```bash
# 1. Create hotfix from master
git checkout master
git pull origin master
git checkout -b fix/critical-auth-bug

# 2. Fix and commit
git add .
git commit -m "fix(auth): resolve authentication bypass vulnerability"

# 3. Push and create PR
git push -u origin fix/critical-auth-bug

# 4. Merge to master and develop
# After approval and testing
```

### Keeping Branch Updated
```bash
# Rebase feature branch with latest master
git checkout master
git pull origin master
git checkout feat/my-feature
git rebase master

# Resolve conflicts if any
git add .
git rebase --continue

# Force push (only for feature branches)
git push --force-with-lease origin feat/my-feature
```

## 📋 Best Practices

### Do's
1. **Commit early and often**
2. **Write meaningful commit messages**
3. **Keep commits focused**
4. **Review your own PR first**
5. **Test before pushing**
6. **Keep master/main stable**
7. **Use feature flags for incomplete features**
8. **Document breaking changes**

### Don'ts
1. **Don't commit directly to master**
2. **Don't commit sensitive data**
3. **Don't commit generated files**
4. **Don't rewrite published history**
5. **Don't merge without review**
6. **Don't ignore CI failures**
7. **Don't commit commented code**
8. **Don't commit debug statements**

## 🏷️ Release Process

### Version Tagging
```bash
# Semantic versioning: MAJOR.MINOR.PATCH
git tag -a v1.2.0 -m "Release version 1.2.0"
git push origin v1.2.0

# Tag naming
v1.0.0     # Major release
v1.1.0     # Minor release
v1.1.1     # Patch release
v2.0.0-rc1 # Release candidate
v2.0.0-beta.1 # Beta release
```

### Release Notes Template
```markdown
# Release v1.2.0

## 🎉 New Features
- Team management system
- Notification framework
- OAuth provider configuration

## 🐛 Bug Fixes
- Fixed authentication timeout issue
- Resolved duplicate notification problem

## 📚 Documentation
- Updated API documentation
- Added deployment guide

## ⚠️ Breaking Changes
- Team API response format changed

## 🔄 Migration Guide
Instructions for migrating from v1.1.x
```

## 🛠️ Git Hooks

### Pre-commit Hook
```bash
#!/bin/sh
# .git/hooks/pre-commit

# Run linting
npm run lint || exit 1

# Check for debug code
if grep -r "console.log\|debugger" --include="*.ts" --include="*.tsx" .; then
    echo "Error: Debug statements found"
    exit 1
fi

# Run type checking
npm run typecheck || exit 1
```

### Commit Message Hook
```bash
#!/bin/sh
# .git/hooks/commit-msg

# Check commit message format
commit_regex='^(feat|fix|docs|style|refactor|perf|test|chore)(\(.+\))?: .{1,50}'

if ! grep -qE "$commit_regex" "$1"; then
    echo "Invalid commit message format!"
    echo "Format: <type>(<scope>): <subject>"
    exit 1
fi
```

## 🔍 Troubleshooting

### Common Issues

#### Merge Conflicts
```bash
# Resolve conflicts manually, then:
git add .
git commit -m "fix: resolve merge conflicts"
```

#### Accidentally Committed to Master
```bash
# Create a new branch with your changes
git branch feat/my-feature

# Reset master to origin
git reset --hard origin/master

# Switch to feature branch
git checkout feat/my-feature
```

#### Need to Undo Last Commit
```bash
# Keep changes
git reset --soft HEAD~1

# Discard changes
git reset --hard HEAD~1
```