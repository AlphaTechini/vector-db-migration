# Repository Setup Instructions

## Issue
GitHub CLI is authenticated as `JnrDevClaw` instead of `AlphaTechini`, preventing repo creation.

## Manual Setup Steps

### Option 1: Create via GitHub Web UI (Recommended)

1. **Go to GitHub**: https://github.com/new
2. **Repository name**: `vector-db-migration`
3. **Owner**: AlphaTechini
4. **Visibility**: Public
5. **Initialize**: ❌ Don't initialize (we have local commits)
6. Click **Create repository**

7. **Push existing code**:
```bash
cd /config/.openclaw/workspace/vector-db-migration
git remote add origin https://github.com/AlphaTechini/vector-db-migration.git
GH_TOKEN=$(gh auth token) git push -u origin main
```

### Option 2: Fix GitHub CLI Auth

```bash
# Logout current session
gh auth logout

# Login as AlphaTechini
gh auth login \
  --hostname github.com \
  --git-protocol https \
  --web

# Then create repo
gh repo create vector-db-migration --public --source=. --remote=origin --push
```

### Option 3: Use Personal Access Token

```bash
# Set token explicitly
export GH_TOKEN=ghp_your_actual_token_here

# Create and push
cd /config/.openclaw/workspace/vector-db-migration
git remote add origin https://github.com/AlphaTechini/vector-db-migration.git
git push -u origin main
```

## Next Steps After Repo Creation

1. ✅ Add landing page (`landing/index.html`) to repo
2. ✅ Set up Vercel deployment
3. ⏳ Configure domain (vectormigrate.dev)
4. ⏳ Add waitlist backend (ConvertKit/Resend)
5. ⏳ Begin system design discussion

---

*Created: February 22, 2026*
