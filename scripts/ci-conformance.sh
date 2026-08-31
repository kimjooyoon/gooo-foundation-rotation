#!/usr/bin/env bash
set -Eeuo pipefail

repo_root="$(pwd)"
run_id="${GITHUB_RUN_ID:-unknown}"
work_root="${RUNNER_TEMP:-/tmp}/gooo-foundation-rotation-${run_id}"
evidence="${work_root}/evidence"
mkdir -p "${evidence}/cases"

if [[ -n "$(gofmt -l cmd internal)" ]]; then
  echo "gofmt reported unformatted Go files" >&2
  gofmt -l cmd internal >&2
  exit 1
fi

go vet ./...

test_start="$(date +%s%N)"
/usr/bin/time -f '%e %M' -o "${work_root}/test.time" go test -count=1 -json ./... > "${work_root}/go-test.json"
test_end="$(date +%s%N)"

/usr/bin/time -f '%e %M' -o "${work_root}/build.time" bash -c \
  'go build -trimpath -o "$1/verifier" ./cmd/gooo-foundation-rotation && go build -trimpath -o "$1/consumer" ./cmd/gooo-foundation-consumer' \
  bash "${work_root}"

/usr/bin/time -f '%e %M' -o "${work_root}/compile.time" \
  "${work_root}/verifier" compile --source semantic/foundation-authorization.gooo --output "${evidence}/semantic-ir.json"

conformance_start="$(date +%s%N)"
/usr/bin/time -f '%e %M' -o "${work_root}/conformance.time" \
  "${work_root}/verifier" conformance --cases-dir fixtures/cases --graph semantic/foundation-authorization.gooo \
  --trust fixtures/trusted-issuers.json --output-dir "${evidence}/cases"
conformance_end="$(date +%s%N)"

"${work_root}/consumer" --input-dir "${evidence}/cases" --output "${evidence}/independent-consumer-receipt.json"

index="${evidence}/cases/conformance-index.json"
test -f "${index}"
test "$(jq '.cases | length' "${index}")" -eq 12
test "$(jq '.states.CLOSED' "${index}")" -ge 3
test "$(jq '.states.UNKNOWN' "${index}")" -ge 3
test "$(jq '.states.REFUTED' "${index}")" -ge 3
test "$(jq '.proof_choices.FOUNDATION' "${index}")" -eq 4
test "$(jq '.proof_choices.COHERENCE' "${index}")" -eq 4
test "$(jq '.proof_choices.REGRESSION' "${index}")" -eq 4
test "$(jq '.indicators.NORMAL' "${index}")" -eq 4
test "$(jq '.indicators.UNKNOWN' "${index}")" -eq 4
test "$(jq '.indicators.REFUTED' "${index}")" -eq 4
test "$(jq '[.cases[] | select(.decision == "REFUTED")] | length' "${index}")" -ge 3
test "$(jq '[.cases[] | select(.decision == "UNKNOWN")] | length' "${index}")" -ge 3

while IFS= read -r result; do
  jq -e '.decision == "CLOSED" or .decision == "UNKNOWN" or .decision == "REFUTED"' "${result}" >/dev/null
  jq -e 'all(.unknowns[]?; .stage != "" and .step != "" and .reason != "" and .unknown_class != "" and .next_operation != "" and (.blocked_by != null))' "${result}" >/dev/null
done < <(find "${evidence}/cases" -mindepth 2 -maxdepth 2 -type f -name verification.json -print | sort)

read -r build_seconds build_rss < "${work_root}/build.time"
read -r compile_seconds compile_rss < "${work_root}/compile.time"
read -r test_seconds test_rss < "${work_root}/test.time"
read -r conformance_seconds conformance_rss < "${work_root}/conformance.time"
ms_from_seconds() { awk -v value="$1" 'BEGIN { printf "%d", (value * 1000) + 0.5 }'; }
build_wall_ms="$(ms_from_seconds "${build_seconds}")"
compile_wall_ms="$(ms_from_seconds "${compile_seconds}")"
test_wall_ms="$(( (test_end - test_start) / 1000000 ))"
conformance_wall_ms="$(( (conformance_end - conformance_start) / 1000000 ))"
peak_rss_kib="$(awk -v a="${build_rss}" -v b="${compile_rss}" -v c="${test_rss}" -v d="${conformance_rss}" 'BEGIN { m=a; if (b>m) m=b; if (c>m) m=c; if (d>m) m=d; print m }')"

tests_total="$(jq -s '[.[] | select(.Action == "run" and .Test != null)] | length' "${work_root}/go-test.json")"
tests_executed="$(jq -s '[.[] | select((.Action == "pass" or .Action == "fail") and .Test != null)] | length' "${work_root}/go-test.json")"
tests_reused="$(jq -s '[.[] | select(.Action == "output" and .Test != null and ((.Output // "") | contains("(cached)")))] | length' "${work_root}/go-test.json")"
tests_failed="$(jq -s '[.[] | select(.Action == "fail" and .Test != null)] | length' "${work_root}/go-test.json")"
tests_skipped="$(jq -s '[.[] | select(.Action == "skip" and .Test != null)] | length' "${work_root}/go-test.json")"
tests_unknown="$(( tests_total - tests_executed - tests_skipped ))"
if (( tests_unknown < 0 )); then tests_unknown=0; fi

go_files="$(find . -type f -name '*.go' ! -path './.git/*' | wc -l | tr -d ' ')"
go_lines="$(find . -type f -name '*.go' ! -path './.git/*' -print0 | xargs -0 -r awk '{ n++ } END { print n + 0 }')"
gooo_files="$(find . -type f -name '*.gooo' ! -path './.git/*' | wc -l | tr -d ' ')"
gooo_lines="$(find . -type f -name '*.gooo' ! -path './.git/*' -print0 | xargs -0 -r awk '{ n++ } END { print n + 0 }')"
regular_files="$(find . -type f ! -path './.git' ! -path './.git/*' ! -path './README.md' | wc -l | tr -d ' ')"
directories="$(find . -mindepth 1 -type d ! -path './.git' ! -path './.git/*' | wc -l | tr -d ' ')"
outputs="$(find "${evidence}" -type f | wc -l | tr -d ' ')"
bytes="$(find "${evidence}" -type f -printf '%s\n' | awk '{ n += $1 } END { print n + 0 }')"

jq -S -n \
  --arg schema 'gooo/foundation-authorization/ci-report/v1' \
  --arg run_id "${run_id}" \
  --argjson compile_wall_ms "${compile_wall_ms}" \
  --argjson build_wall_ms "${build_wall_ms}" \
  --argjson test_wall_ms "${test_wall_ms}" \
  --argjson conformance_wall_ms "${conformance_wall_ms}" \
  --argjson peak_rss_kib "${peak_rss_kib}" \
  --argjson tests_total "${tests_total}" \
  --argjson tests_executed "${tests_executed}" \
  --argjson tests_reused "${tests_reused}" \
  --argjson tests_failed "${tests_failed}" \
  --argjson tests_unknown "${tests_unknown}" \
  --argjson tests_skipped "${tests_skipped}" \
  --argjson go_files "${go_files}" \
  --argjson go_lines "${go_lines}" \
  --argjson gooo_files "${gooo_files}" \
  --argjson gooo_lines "${gooo_lines}" \
  --argjson regular_files "${regular_files}" \
  --argjson directories "${directories}" \
  --argjson outputs "${outputs}" \
  --argjson bytes "${bytes}" \
  --slurpfile index "${index}" \
  '{schema:$schema,ci_run_id:$run_id,scope:{verifier_protocol:"CLOSED",live_issuer:"UNKNOWN",meta_ontology_go_pr_619_integration:"UNKNOWN"},precedence:["REFUTED","UNKNOWN","CLOSED"],denominator:{cells:12,cases:12,proof_choices:$index[0].proof_choices,indicators:$index[0].indicators},corpus:{states:$index[0].states,receipts:$index[0].receipts,rotations:$index[0].rotations,replays:$index[0].replays,revocations:$index[0].revocations},inventory:{go_files:$go_files,go_lines:$go_lines,gooo_files:$gooo_files,gooo_lines:$gooo_lines,regular_files_root_readme_excluded:$regular_files,directories:$directories},outputs:{files:$outputs,bytes:$bytes},timing:{compile_wall_ms:$compile_wall_ms,build_wall_ms:$build_wall_ms,test_wall_ms:$test_wall_ms,conformance_wall_ms:$conformance_wall_ms,peak_rss_kib:$peak_rss_kib},tests:{total:$tests_total,executed:$tests_executed,reused:$tests_reused,failed:$tests_failed,unknown:$tests_unknown,skipped:$tests_skipped},authority:{repository_writes:0,pull_request_creations:0,merge_operations:0,local_test_executions:0,cross_project_required_gates:0},bootstrap:{mode:"direct-main",workflow_present_at_bootstrap:false,direct_main_commits:2,post_ci_bootstrap_direct_main:0,correction_count:1}}' > "${evidence}/ci-report.json"

cat > "${evidence}/human-report.md" <<EOF
# gooo-foundation-rotation CI human report

- decision: CLOSED for verifier/protocol scope only
- live issuer: UNKNOWN
- meta-ontology-go PR #619 integration: UNKNOWN
- precedence: REFUTED > UNKNOWN > CLOSED
- denominator: cells 12, cases 12
- proof choices: FOUNDATION $(jq '.proof_choices.FOUNDATION' "${index}"), COHERENCE $(jq '.proof_choices.COHERENCE' "${index}"), REGRESSION $(jq '.proof_choices.REGRESSION' "${index}")
- indicators: NORMAL $(jq '.indicators.NORMAL' "${index}"), UNKNOWN $(jq '.indicators.UNKNOWN' "${index}"), REFUTED $(jq '.indicators.REFUTED' "${index}")
- receipts: $(jq '.receipts' "${index}"); rotations: $(jq '.rotations' "${index}"); replays: $(jq '.replays' "${index}"); revocations: $(jq '.revocations' "${index}")
- tests: total ${tests_total}, executed ${tests_executed}, reused ${tests_reused}, failed ${tests_failed}, unknown ${tests_unknown}
- compile/build/test/conformance wall_ms: ${compile_wall_ms} / ${build_wall_ms} / ${test_wall_ms} / ${conformance_wall_ms}
- peak_rss_kib: ${peak_rss_kib}
- inventory: Go files/physical lines ${go_files}/${go_lines}, Gooo files/physical lines ${gooo_files}/${gooo_lines}, files excluding root README ${regular_files}, dirs ${directories}
- outputs/files/bytes: ${outputs} / ${bytes}
- authority writes/PRs/merges/local-tests/cross-project-gates: 0/0/0/0/0
- bootstrap correction count: 1; future repository-bootstrap capability gap retained
- bootstrap direct-main commits: 2; post-CI-bootstrap direct-main commits: 0

The regression corpus includes known REFUTED cases. Those cases are evidence
that the guard rejects replay, mismatch, revocation, and self-authorization;
they do not lower the conformance decision to UNKNOWN.
EOF

echo "evidence=${evidence}"
