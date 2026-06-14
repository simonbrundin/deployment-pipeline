# Deployment Pipeline

Mall för CI/CD-pipelines med Dagger, Argo Workflows och GitHub Actions.

## Pipeline-steg

| Steg | Beskrivning                                       |
| ---- | ------------------------------------------------- |
| Bygg | Bygger Docker-image från Dockerfile eller källkod |
| Test | Kör enhetstester (Go, JavaScript, Java)           |
| Push | Pushar image till registry (stödjer multi-arch)   |

## Stödda språk

- Go
- JavaScript (via Bun)
- Java (via Maven)

## Användning

### GitHub Actions

Kopiera `workflows/github-actions/dagger.yaml` till din repo.

### Argo Workflows

Använd `ClusterWorkflowTemplate` från
`workflows/argo-workflows/deployment-pipeline.yaml`.

### Argos Events

Trigger från `events/argo-events/commit.yaml`.

## Struktur

```
├── dagger-modules/pipeline/   # Dagger Go-moduler
├── workflows/                 # Pipeline-definitioner
└── events/                    # Event-triggers
```
