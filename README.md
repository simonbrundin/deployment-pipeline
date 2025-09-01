# Deployment Pipeline

Här finns en Deployment Pipeline jag kan använda som mall för alla applications
jag utvecklar.

## Faser

### Commit Phase

Paketerar produkten och ger utvecklaren tummen upp att påbörja nästa förändring.

**Nödvändiga komponenter:**

- Applikationsrepository
- Commit-workflow
- Artifaktrepository
- Program som lyssnar efter commits i repositoryt och kör commit-workflow

**Commit-workflow:**

1. Sätt upp testmiljö
2. Bygg applikationen
3. Kör enhetstester
4. Skapa artefakt

### Acceptance Phase

### Production Phase

### Production Phase
