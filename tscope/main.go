package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

type ScopeNode struct {
	Identifier            string   `json:"identifier"`
	DisplayName           string   `json:"display_name"`
	CvssScore             string   `json:"cvss_score"`
	EligibleForBounty     bool     `json:"eligible_for_bounty"`
	EligibleForSubmission bool     `json:"eligible_for_submission"`
	AsmSystemTags         []string `json:"asm_system_tags"`
	TotalResolvedReports  int      `json:"total_resolved_reports"`
}

type GraphQLResponse struct {
	Data struct {
		Team struct {
			StructuredScopesSearch struct {
				Nodes []ScopeNode `json:"nodes"`
			} `json:"structured_scopes_search"`
		} `json:"team"`
	} `json:"data"`
}

func main() {
	handle := flag.String("p", "", "HackeOne program handle name")
	cookie := flag.String("c", "", "HackerOne session cookie (__Host-session value)")
	csrf := flag.String("t", "", "HackerOne csrf token (x-csrf-token)")
	flag.Parse()

	if cookie == nil || *cookie == "" {
		*cookie = os.Getenv("H1_COOKIE")
	}
	if csrf == nil || *csrf == "" {
		*csrf = os.Getenv("H1_CSRF")
	}

	if *handle == "" || *cookie == "" || *csrf == "" {
		fmt.Println("Usage: tscope -h <handle> -c <cookie_valu> -t <csrf_value>")
		os.Exit(1)
	}

	query := map[string]any{
		"operationName": "PolicySearchStructuredScopesQuery",
		"variables": map[string]any{
			"handle":                *handle,
			"searchString":          "",
			"eligibleForSubmission": nil,
			"eligibleForBounty":     nil,
			"asmTagIds":             []any{},
			"assetTypes":            []any{},
			"from":                  0,
			"size":                  100,
			"sort": map[string]any{
				"field":     "cvss_score",
				"direction": "DESC",
			},
			"product_area":    "h1_assets",
			"product_feature": "policy_scopes",
		},
		"query": `query PolicySearchStructuredScopesQuery($handle: String!, $searchString: String, $eligibleForSubmission: Boolean, $eligibleForBounty: Boolean, $minSeverityScore: SeverityRatingEnum, $asmTagIds: [Int], $assetTypes: [StructuredScopeAssetTypeEnum!], $from: Int, $size: Int, $sort: SortInput) {
  team(handle: $handle) {
    id
    team_display_options {
      show_total_reports_per_asset
      __typename
    }
    structured_scopes_search(
      search_string: $searchString
      eligible_for_submission: $eligibleForSubmission
      eligible_for_bounty: $eligibleForBounty
      min_severity_score: $minSeverityScore
      asm_tag_ids: $asmTagIds
      asset_types: $assetTypes
      from: $from
      size: $size
      sort: $sort
    ) {
      nodes {
        ... on StructuredScopeDocument {
          id
          ...PolicyScopeStructuredScopeDocument
          __typename
        }
        __typename
      }
      pageInfo {
        startCursor
        hasPreviousPage
        endCursor
        hasNextPage
        __typename
      }
      total_count
      __typename
    }
    __typename
  }
}

fragment PolicyScopeStructuredScopeDocument on StructuredScopeDocument {
  id
  identifier
  display_name
  instruction
  cvss_score
  eligible_for_bounty
  eligible_for_submission
  asm_system_tags
  created_at
  updated_at
  total_resolved_reports
  attachments {
    id
    file_name
    file_size
    content_type
    expiring_url
    __typename
  }
  __typename
}
`,
	}

	body, _ := json.Marshal(query)
	req, err := http.NewRequest("POST", "https://hackerone.com/graphql", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error HTTP Request: ", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "__Host-session="+*cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("x-csrf-token", *csrf)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Reuqest error: ", err)
		os.Exit(1)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		fmt.Printf("[!] Server returned status %d\n", res.StatusCode)
		io.Copy(os.Stdout, res.Body)
		os.Exit(1)
	}

	var gr GraphQLResponse
	if err := json.NewDecoder(res.Body).Decode(&gr); err != nil {
		fmt.Println("Decode error: ", err)
		os.Exit(1)
	}

	var filtered []ScopeNode
	for _, node := range gr.Data.Team.StructuredScopesSearch.Nodes {
		if node.EligibleForBounty && node.EligibleForSubmission {
			filtered = append(filtered, node)
		}
	}

	if len(filtered) == 0 {
		fmt.Println("[!] No eligible scope found or invalid session.")
		os.Exit(1)
	}

	out, _ := json.MarshalIndent(filtered, "", "  ")
	fmt.Println(string(out))
}
