package serve

// HTML template for the dashboard
const dashboardTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MXLint Dashboard</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        h1, h2, h3 {
            color: #0066cc;
        }
        .header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 1px solid #eee;
        }
        .timestamp {
            font-size: 0.9em;
            color: #666;
        }
        .rule {
            background-color: #f9f9f9;
            border-radius: 5px;
            padding: 15px;
            margin-bottom: 20px;
            border-left: 4px solid #0066cc;
        }
        .rule-header {
            display: flex;
            justify-content: space-between;
            margin-bottom: 10px;
        }
        .rule-title {
            font-weight: bold;
            margin: 0;
        }
        .rule-meta {
            display: flex;
            gap: 15px;
            font-size: 0.9em;
        }
        .severity {
            font-weight: bold;
        }
        .severity-HIGH {
            color: #d73a49;
        }
        .severity-MEDIUM {
            color: #e36209;
        }
        .severity-LOW {
            color: #6a737d;
        }
        .testcase {
            padding: 10px;
            margin: 5px 0;
            border-radius: 3px;
        }
        .testcase-pass {
            background-color: #f0fff4;
            border-left: 3px solid #22863a;
        }
        .testcase-fail {
            background-color: #fff5f5;
            border-left: 3px solid #d73a49;
        }
        .testcase-skip {
            background-color: #f8f8f8;
            border-left: 3px solid #6a737d;
        }
        .testcase-header {
            display: flex;
            justify-content: space-between;
        }
        .failure-message {
            background-color: #fff5f5;
            border-radius: 3px;
            padding: 10px;
            margin-top: 5px;
            font-family: monospace;
            white-space: pre-wrap;
        }
        .refresh-button {
            background-color: #0066cc;
            color: white;
            border: none;
            padding: 8px 16px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
        }
        .refresh-button:hover {
            background-color: #0055aa;
        }
        .summary {
            display: flex;
            gap: 20px;
            margin-bottom: 20px;
        }
        .summary-item {
            padding: 10px 15px;
            border-radius: 5px;
            text-align: center;
            cursor: pointer;
            transition: transform 0.1s ease;
            user-select: none;
        }
        .summary-item:hover {
            transform: translateY(-2px);
        }
        .summary-item.active {
            box-shadow: 0 0 0 2px #0066cc;
        }
        .summary-total {
            background-color: #f0f7ff;
            border: 1px solid #cce5ff;
        }
        .summary-failures {
            background-color: #fff5f5;
            border: 1px solid #ffdce0;
        }
        .summary-skipped {
            background-color: #f8f8f8;
            border: 1px solid #e1e4e8;
        }
        .summary-number {
            font-size: 1.5em;
            font-weight: bold;
        }
        .auto-refresh {
            display: flex;
            align-items: center;
            gap: 10px;
            font-size: 0.9em;
        }
        .hidden {
            display: none !important;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>MXLint Dashboard</h1>
        <div>
            <div class="timestamp">Last updated: {{.Timestamp.Format "Jan 02, 2006 15:04:05"}}</div>
            <div class="auto-refresh">
                <input type="checkbox" id="auto-refresh" checked>
                <label for="auto-refresh">Auto-refresh (10s)</label>
                <button class="refresh-button" onclick="location.reload()">Refresh Now</button>
            </div>
        </div>
    </div>

    {{if .Error}}
    <div class="failure-message">{{.Error}}</div>
    {{else}}
    {{with .Results}}
    <div class="summary">
        {{$totalTests := 0}}
        {{$totalFailures := 0}}
        {{$totalSkipped := 0}}
        {{range .Testsuites}}
            {{$totalTests = add $totalTests .Tests}}
            {{$totalFailures = add $totalFailures .Failures}}
            {{$totalSkipped = add $totalSkipped .Skipped}}
        {{end}}
        <div id="filter-all" class="summary-item summary-total" onclick="filterResults('all')">
            <div class="summary-number">{{$totalTests}}</div>
            <div>Total Tests</div>
        </div>
        <div id="filter-failures" class="summary-item summary-failures active" onclick="filterResults('failures')">
            <div class="summary-number">{{$totalFailures}}</div>
            <div>Failures</div>
        </div>
        <div id="filter-skipped" class="summary-item summary-skipped" onclick="filterResults('skipped')">
            <div class="summary-number">{{$totalSkipped}}</div>
            <div>Skipped</div>
        </div>
    </div>

    {{range .Rules}}
    <div class="rule">
        <div class="rule-header">
            <h3 class="rule-title">{{.Title}}</h3>
            <div class="rule-meta">
                <div><span class="severity severity-{{.Severity}}">{{.Severity}}</span></div>
                <div>{{.Category}}</div>
                <div>Rule #{{.RuleNumber}}</div>
            </div>
        </div>
        <p>{{.Description}}</p>
        <p><strong>Remediation:</strong> {{.Remediation}}</p>
        
        {{$rulePath := .Path}}
        {{with $ := $.Results}}
            {{range $testsuite := .Testsuites}}
                {{if eq $testsuite.Name $rulePath}}
                    <h4>Test Results</h4>
                    {{range $testsuite.Testcases}}
                        {{if .Failure}}
                            <div class="testcase testcase-fail result-item result-failure">
                                <div class="testcase-header">
                                    <div>❌ {{.Name}}</div>
                                    <div>{{printf "%.3fs" .Time}}</div>
                                </div>
                                <div class="failure-message">{{.Failure.Message}}</div>
                            </div>
                        {{else if .Skipped}}
                            <div class="testcase testcase-skip result-item result-skipped">
                                <div class="testcase-header">
                                    <div>⏭️ {{.Name}}</div>
                                    <div>Skipped: {{.Skipped.Message}}</div>
                                </div>
                            </div>
                        {{else}}
                            <div class="testcase testcase-pass result-item result-pass">
                                <div class="testcase-header">
                                    <div>✅ {{.Name}}</div>
                                    <div>{{printf "%.3fs" .Time}}</div>
                                </div>
                            </div>
                        {{end}}
                    {{end}}
                {{end}}
            {{end}}
        {{end}}
    </div>
    {{end}}
    {{end}}
    {{end}}

    <script>
        // Auto-refresh functionality
        const checkbox = document.getElementById('auto-refresh');
        let refreshInterval;

        function startAutoRefresh() {
            refreshInterval = setInterval(() => {
                if (checkbox.checked) {
                    location.reload();
                }
            }, 10000); // 10 seconds
        }

        checkbox.addEventListener('change', () => {
            if (checkbox.checked) {
                startAutoRefresh();
            } else {
                clearInterval(refreshInterval);
            }
        });

        // Start auto-refresh on page load
        startAutoRefresh();
        
        // Filter functionality
        function filterResults(filter) {
            // Update active state of filter buttons
            document.getElementById('filter-all').classList.remove('active');
            document.getElementById('filter-failures').classList.remove('active');
            document.getElementById('filter-skipped').classList.remove('active');
            document.getElementById('filter-' + filter).classList.add('active');
            
            // Get all result items
            const resultItems = document.querySelectorAll('.result-item');
            
            // Show/hide based on filter
            resultItems.forEach(item => {
                if (filter === 'all') {
                    item.classList.remove('hidden');
                } else if (filter === 'failures') {
                    if (item.classList.contains('result-failure')) {
                        item.classList.remove('hidden');
                    } else {
                        item.classList.add('hidden');
                    }
                } else if (filter === 'skipped') {
                    if (item.classList.contains('result-skipped')) {
                        item.classList.remove('hidden');
                    } else {
                        item.classList.add('hidden');
                    }
                }
            });
            
            // Handle empty rules (hide rules with no visible test cases)
            const rules = document.querySelectorAll('.rule');
            rules.forEach(rule => {
                const visibleTests = rule.querySelectorAll('.result-item:not(.hidden)');
                if (visibleTests.length === 0) {
                    rule.classList.add('hidden');
                } else {
                    rule.classList.remove('hidden');
                }
            });
        }
        
        // Initialize with failures filter on page load
        document.addEventListener('DOMContentLoaded', function() {
            filterResults('failures');
        });
    </script>
</body>
</html>
`
