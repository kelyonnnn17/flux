## Plan: Graph-Based Best-Effort Document Routing

Replace direct-only document validation/execution with a weighted route planner that can choose multi-step pipelines, while keeping current image/audio/data behavior stable. The implementation will add planner metadata + route warnings, enforce strict explicit-engine rules, and introduce PDF input best-effort via text extraction first.

**Steps**
1. Phase 1 - Build planner foundation in [internal/engine/engine.go](internal/engine/engine.go) plus a new planner module.
2. Model graph as format nodes and conversion edges with engine/tool, from/to, cost, and fidelity/warning metadata.
3. Implement weighted shortest path selection (Dijkstra) with deterministic tie-breaks: direct first, then fewer hops, then higher fidelity.
4. Add route metadata types so planning returns steps, total cost, and warnings.
5. Phase 2 - Upgrade validation.
6. Rework validation wrappers in [internal/engine/engine.go](internal/engine/engine.go) so CanConvert and CanEngineConvert validate full routes, not only direct pairs.
7. Enforce strict explicit engine behavior: if forced engine cannot execute all required steps, fail with actionable message and suggest engine auto.
8. Keep non-document conversions behaviorally unchanged unless needed as explicit planner intermediates.
9. Phase 3 - Execute multi-hop routes.
10. Insert pipeline execution in CLI flow at [cmd/convert.go](cmd/convert.go) using temp intermediates, per-step execution, cleanup, and final atomic output behavior.
11. Mirror the same route execution logic in library flow at [pkg/flux/convert.go](pkg/flux/convert.go) for parity.
12. Preserve spinner/UI behavior and add concise warning/route summary output using existing formatting conventions.
13. Phase 4 - PDF input best-effort.
14. Add planner edges for PDF source handling with preferred text extraction route (pdftotext first), then document routing to target via pandoc-compatible steps.
15. Add robust tool availability checks, timeout-safe invocation, and clear install/action hints when extraction tools are missing.
16. If all PDF routes fail, return clear route-attempt summary and actionable next steps (no partial outputs by default).
17. Phase 5 - Tests and docs.
18. Expand planner/validation tests in [internal/engine/engine_test.go](internal/engine/engine_test.go) for direct preference, fallback routes, forced-engine failures, and warning metadata.
19. Expand CLI/library tests in [cmd/convert_test.go](cmd/convert_test.go) and [pkg/flux/convert_test.go](pkg/flux/convert_test.go) for multi-step execution and messaging.
20. Add integration coverage in [tests/integration/integration_test.go](tests/integration/integration_test.go) for document fallback pipelines and PDF best-effort success/failure (with tool-aware skips).
21. Update user-facing behavior docs and caveats in [README.md](README.md), [REQUIREMENTS.md](REQUIREMENTS.md), and output messaging in [cmd/list_formats.go](cmd/list_formats.go).

**Relevant files**
- [internal/engine/engine.go](internal/engine/engine.go): current capability/validation/routing hooks to planner-back.
- [internal/engine/pandoc.go](internal/engine/pandoc.go): document engine execution path.
- [internal/engine/imagemagick.go](internal/engine/imagemagick.go): compatibility boundaries for intermediates.
- [cmd/convert.go](cmd/convert.go): CLI validation/execution integration point.
- [pkg/flux/convert.go](pkg/flux/convert.go): library conversion parity.
- [cmd/list_formats.go](cmd/list_formats.go): format UX notes and caveats.
- [internal/engine/engine_test.go](internal/engine/engine_test.go): planner and validation unit tests.
- [cmd/convert_test.go](cmd/convert_test.go): CLI multi-hop behavior tests.
- [pkg/flux/convert_test.go](pkg/flux/convert_test.go): package behavior tests.
- [tests/integration/integration_test.go](tests/integration/integration_test.go): end-to-end route coverage.
- [README.md](README.md): user-facing behavior/caveats.
- [REQUIREMENTS.md](REQUIREMENTS.md): requirement-level updates.

**Verification**
1. Run targeted engine tests for planner/path selection, forced-engine constraints, and warning emission.
2. Run CLI/library tests for route execution parity and explicit-engine failure guidance.
3. Run integration tests for multi-hop document routes and PDF best-effort with tool-present/tool-missing scenarios.
4. Manually validate representative flows: docx to pdf, rst to docx via md, pdf to docx best-effort, and impossible route messaging.
5. Re-run image/audio/data regressions to confirm unchanged behavior.

**Decisions captured**
- PDF best-effort default: pdftotext-first route.
- Default route bias: direct over multi-hop, then fewer hops, then fidelity.
- Explicit engine behavior: strict end-to-end compatibility required; otherwise fail with suggestion to use auto.
- OCR is excluded from default scope for this phase and can be added later as an optional extension.

I have also persisted this complete plan in session memory so it is ready for handoff execution.
