# Deployment Pipeline

Reusable CI/CD pipelines för Dagger, anropbara från andra GitHub-repos.

## Användning i andra repos

```yaml
# .github/workflows/ci.yaml
name: CI

on: push

jobs:
  commit:
    uses: simonbrundin/deployment-pipeline/.github/workflows/commit-phase.yaml@v1
    with:
      source-dir: . # Valfri, default: .
      image-name: "" # Valfri, default: repo-namn

  acceptance:
    needs: commit
    uses: simonbrundin/deployment-pipeline/.github/workflows/acceptance-phase.yaml@v1
    with:
      source-dir: . # Valfri, default: .
```

## Krävs i anropande repo

**Secrets:**

- `DAGGER_CLOUD_TOKEN`

**Variables:**

- `REGISTRY_ADDRESS`
- `REGISTRY_USERNAME`
- `REGISTRY_PASSWORD`

## Struktur

```
.github/workflows/
├── ci.yaml              # Kombinerad pipeline (för test)
├── commit-phase.yaml    # Reusable: bygg + push
└── acceptance-phase.yaml # Reusable: test + deploy
```

## Versionering

Använd taggar (`@v1`, `@v2`) för att låsa version. Se till att taggen finns i
detta repo innan du uppdaterar referenser i produktionsrepos.
