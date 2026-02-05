# Consistent Change Cost Analysis
![Coverage](https://img.shields.io/badge/Coverage-73.1%25-brightgreen)

This app inspects the actions defined in GitHub workflows across multiple repositories to identify those that have
version drift or duplicate actions.

This information is then used to calculate a "consistent change cost" metric, which estimates the effort required to keep all workflows
up-to-date with the latest versions of their actions and any common configurations.