package llm

import (
	"bytes"
	"encoding/json"
	"text/template"
)

const systemPromptTemplate = `あなたはMinecraft Bedrock Editionサーバーのテストシナリオを実行するAIエージェントです。
ユーザーが提供する自然言語のシナリオを、実行可能なステップに変換してください。

## 利用可能なアクション
{{range .Actions}}
- {{.Name}}: {{.Description}}
{{- if .Parameters}}
  パラメータ:
{{- range .Parameters}}
    - {{.Name}} ({{.Type}}{{if .Required}}, 必須{{end}}): {{.Description}}{{if .Default}} (デフォルト: {{.Default}}){{end}}
{{- end}}
{{- end}}
{{end}}

## 利用可能なアサーション
{{range .Assertions}}
- {{.Name}}: {{.Description}}
{{- if .Parameters}}
  パラメータ:
{{- range .Parameters}}
    - {{.Name}} ({{.Type}}{{if .Required}}, 必須{{end}}): {{.Description}}{{if .Default}} (デフォルト: {{.Default}}){{end}}
{{- end}}
{{- end}}
{{end}}

## 出力形式
シナリオを解析し、以下のJSON形式で出力してください。必ずJSONのみを出力し、他のテキストは含めないでください。

{
  "steps": [
    {
      "action": "アクション名またはアサーション名",
      "description": "ステップの説明（日本語）",
      "params": {
        "パラメータ名": "値"
      }
    }
  ]
}

## 注意事項
- 各ステップは上記のアクションまたはアサーションのいずれかを使用してください
- パラメータは適切な型で指定してください（文字列は"で囲む、数値はそのまま）
- 待機時間（duration）は "2s", "500ms", "1m" などの形式で指定してください
- シナリオの意図を正確に理解し、適切なステップに変換してください
- 接続が必要な場合は最初にconnectアクションを含めてください
`

const userPromptTemplate = `以下のシナリオを実行可能なステップに変換してください：

{{.ScenarioText}}
`

type promptData struct {
	Actions    []ActionDefinition
	Assertions []AssertionDefinition
}

type userPromptData struct {
	ScenarioText string
}

// BuildSystemPrompt builds the system prompt with available actions and assertions
func BuildSystemPrompt(ctx *ScenarioContext) (string, error) {
	tmpl, err := template.New("system").Parse(systemPromptTemplate)
	if err != nil {
		return "", err
	}

	data := promptData{
		Actions:    ctx.AvailableActions,
		Assertions: ctx.AvailableAssertions,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// BuildUserPrompt builds the user prompt with the scenario text
func BuildUserPrompt(scenarioText string) (string, error) {
	tmpl, err := template.New("user").Parse(userPromptTemplate)
	if err != nil {
		return "", err
	}

	data := userPromptData{
		ScenarioText: scenarioText,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// BuildSummaryPrompt builds a prompt for generating a test summary
func BuildSummaryPrompt(results *SummaryInput) string {
	jsonData, _ := json.MarshalIndent(results, "", "  ")
	return `以下のテスト結果を分析し、簡潔なサマリーを日本語で生成してください。
重要なポイント、成功/失敗の概要、失敗したテストがあればその原因を含めてください。

テスト結果:
` + string(jsonData)
}

// ExtractJSONFromResponse extracts JSON from the LLM response
// It handles cases where the response might contain markdown code blocks
func ExtractJSONFromResponse(response string) ([]ScenarioStep, error) {
	// Try to find JSON in code blocks first
	jsonStr := response

	// Remove markdown code blocks if present
	if start := bytes.Index([]byte(response), []byte("```json")); start != -1 {
		end := bytes.Index([]byte(response[start+7:]), []byte("```"))
		if end != -1 {
			jsonStr = response[start+7 : start+7+end]
		}
	} else if start := bytes.Index([]byte(response), []byte("```")); start != -1 {
		end := bytes.Index([]byte(response[start+3:]), []byte("```"))
		if end != -1 {
			jsonStr = response[start+3 : start+3+end]
		}
	}

	// Parse the JSON
	var result ParseResponse
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		// Try parsing as an array directly
		var steps []ScenarioStep
		if err2 := json.Unmarshal([]byte(jsonStr), &steps); err2 != nil {
			return nil, err
		}
		return steps, nil
	}

	return result.Steps, nil
}
