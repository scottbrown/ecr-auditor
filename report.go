package ecrauditor

import (
	"fmt"
	"html/template"
	"io"
	"time"
)

const reportTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ECR Audit Report - {{.Region}}</title>
    <style>
        :root {
            --primary: #232f3e;
            --secondary: #ff9900;
            --success: #1d8102;
            --warning: #ff9900;
            --danger: #d13212;
            --bg: #f8f9fa;
            --card-bg: #ffffff;
            --text: #16191f;
            --text-muted: #687078;
            --border: #d5dbdb;
            --hover-bg: #f1f3f3;
            --images-bg: #f8f9fa;
        }
        [data-theme="dark"] {
            --primary: #1a2332;
            --secondary: #ff9900;
            --success: #2ea043;
            --warning: #ff9900;
            --danger: #f85149;
            --bg: #0d1117;
            --card-bg: #161b22;
            --text: #e6edf3;
            --text-muted: #8b949e;
            --border: #30363d;
            --hover-bg: #21262d;
            --images-bg: #21262d;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--bg);
            color: var(--text);
            line-height: 1.6;
        }
        .header {
            background: var(--primary);
            color: white;
            padding: 2rem;
            text-align: center;
            position: relative;
        }
        .header h1 { font-size: 1.8rem; margin-bottom: 0.5rem; }
        .header .meta { color: #aab7b8; font-size: 0.9rem; }
        .theme-toggle {
            position: absolute;
            top: 1rem;
            right: 1rem;
            background: var(--secondary);
            border: none;
            border-radius: 20px;
            padding: 0.5rem 1rem;
            cursor: pointer;
            font-size: 0.9rem;
            color: var(--primary);
            font-weight: 500;
            transition: opacity 0.2s;
        }
        .theme-toggle:hover { opacity: 0.9; }
        .container { max-width: 1200px; margin: 0 auto; padding: 2rem; }
        .summary-cards {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1rem;
            margin-bottom: 2rem;
        }
        .card {
            background: var(--card-bg);
            border-radius: 8px;
            padding: 1.5rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        .card.stat { text-align: center; }
        .card.stat .value {
            font-size: 2.5rem;
            font-weight: bold;
            color: var(--secondary);
        }
        .card.stat .label {
            color: var(--text-muted);
            font-size: 0.9rem;
        }
        .card.stat.warning .value { color: var(--warning); }
        .card.stat.success .value { color: var(--success); }
        .card.stat.danger .value { color: var(--danger); }
        .section { margin-bottom: 2rem; }
        .section h2 {
            font-size: 1.4rem;
            margin-bottom: 1rem;
            padding-bottom: 0.5rem;
            border-bottom: 2px solid var(--secondary);
        }
        .repo-table {
            width: 100%;
            border-collapse: collapse;
            background: var(--card-bg);
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        .repo-table th, .repo-table td {
            padding: 1rem;
            text-align: left;
            border-bottom: 1px solid var(--border);
        }
        .repo-table th {
            background: var(--primary);
            color: white;
            font-weight: 600;
            cursor: pointer;
            user-select: none;
            white-space: nowrap;
        }
        .repo-table th:hover { background: #1a2633; }
        [data-theme="dark"] .repo-table th:hover { background: #0d1520; }
        .repo-table th .sort-indicator {
            margin-left: 0.5rem;
            opacity: 0.5;
        }
        .repo-table th.sort-asc .sort-indicator,
        .repo-table th.sort-desc .sort-indicator { opacity: 1; }
        .repo-table tr:hover { background: var(--hover-bg); }
        .repo-table tr:last-child td { border-bottom: none; }
        .badge {
            display: inline-block;
            padding: 0.25rem 0.75rem;
            border-radius: 20px;
            font-size: 0.8rem;
            font-weight: 500;
        }
        .badge.success { background: #d4edda; color: #155724; }
        .badge.danger { background: #f8d7da; color: #721c24; }
        [data-theme="dark"] .badge.success { background: #238636; color: #ffffff; }
        [data-theme="dark"] .badge.danger { background: #da3633; color: #ffffff; }
        .expandable { cursor: pointer; }
        .expandable::before { content: '▶ '; font-size: 0.7rem; }
        .expandable.open::before { content: '▼ '; }
        .images-list {
            display: none;
            padding: 1rem;
            background: var(--images-bg);
            margin-top: 0.5rem;
            border-radius: 4px;
            font-size: 0.85rem;
        }
        .images-list.open { display: block; }
        .images-list table { width: 100%; }
        .images-list th, .images-list td {
            padding: 0.5rem;
            text-align: left;
            border-bottom: 1px solid var(--border);
        }
        .images-list th {
            background: transparent;
            color: var(--text);
            font-weight: 600;
            cursor: pointer;
            user-select: none;
            white-space: nowrap;
        }
        .images-list th:hover { background: var(--hover-bg); }
        .images-list th .sort-indicator {
            margin-left: 0.5rem;
            opacity: 0.5;
        }
        .images-list th.sort-asc .sort-indicator,
        .images-list th.sort-desc .sort-indicator { opacity: 1; }
        .tag {
            display: inline-block;
            background: var(--secondary);
            color: var(--primary);
            padding: 0.1rem 0.4rem;
            border-radius: 3px;
            font-size: 0.75rem;
            margin: 0.1rem;
        }
        .lifecycle-policy {
            background: var(--images-bg);
            padding: 1rem;
            border-radius: 4px;
            font-family: monospace;
            font-size: 0.8rem;
            white-space: pre-wrap;
            max-height: 200px;
            overflow: auto;
        }
        .filter-bar {
            margin-bottom: 1rem;
            display: flex;
            gap: 1rem;
            align-items: center;
        }
        .filter-bar input {
            padding: 0.5rem 1rem;
            border: 1px solid var(--border);
            border-radius: 4px;
            font-size: 1rem;
            flex: 1;
            max-width: 400px;
            background: var(--card-bg);
            color: var(--text);
        }
        .filter-bar select {
            padding: 0.5rem 1rem;
            border: 1px solid var(--border);
            border-radius: 4px;
            font-size: 1rem;
            background: var(--card-bg);
            color: var(--text);
        }
        .footer {
            text-align: center;
            padding: 2rem;
            color: var(--text-muted);
            font-size: 0.85rem;
        }
        @media (max-width: 768px) {
            .summary-cards { grid-template-columns: repeat(2, 1fr); }
            .repo-table { font-size: 0.9rem; }
            .repo-table th, .repo-table td { padding: 0.75rem 0.5rem; }
            .theme-toggle { position: static; margin-top: 1rem; }
        }
    </style>
</head>
<body>
    <div class="header">
        <button class="theme-toggle" onclick="toggleTheme()">Toggle Dark Mode</button>
        <h1>ECR Audit Report</h1>
        <div class="meta">
            Region: <strong>{{.Region}}</strong> |
            Profile: <strong>{{.Profile}}</strong> |
            Generated: <strong>{{.CollectedAt.Format "2006-01-02 15:04:05 MST"}}</strong>
        </div>
    </div>

    <div class="container">
        <div class="summary-cards">
            <div class="card stat">
                <div class="value">{{.RepositoryCount}}</div>
                <div class="label">Repositories</div>
            </div>
            <div class="card stat">
                <div class="value">{{.TotalImageCount}}</div>
                <div class="label">Total Images</div>
            </div>
            <div class="card stat success">
                <div class="value">{{.RepositoriesWithLifecyclePolicy}}</div>
                <div class="label">With Lifecycle Policy</div>
            </div>
            <div class="card stat {{if gt .RepositoriesWithoutLifecyclePolicy 0}}danger{{else}}success{{end}}">
                <div class="value">{{.RepositoriesWithoutLifecyclePolicy}}</div>
                <div class="label">Without Lifecycle Policy</div>
            </div>
        </div>

        <div class="section">
            <h2>Repositories</h2>
            <div class="filter-bar">
                <input type="text" id="search" placeholder="Search repositories..." onkeyup="filterTable()">
                <select id="policyFilter" onchange="filterTable()">
                    <option value="all">All Repositories</option>
                    <option value="with">With Lifecycle Policy</option>
                    <option value="without">Without Lifecycle Policy</option>
                </select>
            </div>
            <table class="repo-table" id="repoTable">
                <thead>
                    <tr>
                        <th data-sort="name" onclick="sortTable(event, this, 'name')">Repository Name<span class="sort-indicator">&#x2195;</span></th>
                        <th data-sort="images" onclick="sortTable(event, this, 'images')">Images<span class="sort-indicator">&#x2195;</span></th>
                        <th data-sort="median" onclick="sortTable(event, this, 'median')">Median/Day<span class="sort-indicator">&#x2195;</span></th>
                        <th data-sort="policy" onclick="sortTable(event, this, 'policy')">Lifecycle Policy<span class="sort-indicator">&#x2195;</span></th>
                        <th data-sort="created" onclick="sortTable(event, this, 'created')">Created<span class="sort-indicator">&#x2195;</span></th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Repositories}}
                    <tr data-has-policy="{{if .LifecyclePolicy}}true{{else}}false{{end}}" data-name="{{.Name}}" data-images="{{.ImageCount}}" data-median="{{printf "%.1f" .MedianImagesPerDay}}" data-policy="{{if .LifecyclePolicy}}1{{else}}0{{end}}" data-created="{{.CreatedAt.Unix}}">
                        <td>
                            <span class="expandable" onclick="toggleImages(this)">{{.Name}}</span>
                            <div class="images-list">
                                {{if .Images}}
                                <table class="images-table">
                                    <thead>
                                        <tr>
                                            <th data-sort="tags" onclick="sortImagesTable(event, this, 'tags')">Tags<span class="sort-indicator">&#x2195;</span></th>
                                            <th data-sort="size" onclick="sortImagesTable(event, this, 'size')">Size<span class="sort-indicator">&#x2195;</span></th>
                                            <th data-sort="pushed" class="sort-desc" onclick="sortImagesTable(event, this, 'pushed')">Pushed<span class="sort-indicator">&#x2195;</span></th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {{range .Images}}
                                        <tr data-tags="{{range .Tags}}{{.}} {{end}}" data-size="{{.SizeBytes}}" data-pushed="{{.PushedAt.Unix}}">
                                            <td>
                                                {{if .Tags}}
                                                    {{range .Tags}}<span class="tag">{{.}}</span>{{end}}
                                                {{else}}
                                                    <em>untagged</em>
                                                {{end}}
                                            </td>
                                            <td>{{formatBytes .SizeBytes}}</td>
                                            <td>{{.PushedAt.Format "2006-01-02"}}</td>
                                        </tr>
                                        {{end}}
                                    </tbody>
                                </table>
                                {{else}}
                                <em>No images</em>
                                {{end}}
                            </div>
                        </td>
                        <td>{{.ImageCount}}</td>
                        <td>{{printf "%.1f" .MedianImagesPerDay}}</td>
                        <td>
                            {{if .LifecyclePolicy}}
                                <span class="badge success">Yes</span>
                            {{else}}
                                <span class="badge danger">No</span>
                            {{end}}
                        </td>
                        <td>{{.CreatedAt.Format "2006-01-02"}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>

        {{if .RepositoriesWithoutLifecyclePolicy}}
        <div class="section">
            <h2>Repositories Missing Lifecycle Policies</h2>
            <p style="margin-bottom: 1rem; color: var(--text-muted);">
                The following repositories do not have lifecycle policies configured, which may lead to unnecessary storage costs.
            </p>
            <ul style="list-style: none;">
                {{range .Repositories}}
                {{if not .LifecyclePolicy}}
                <li style="padding: 0.5rem 0; border-bottom: 1px solid var(--border);">
                    <strong>{{.Name}}</strong>
                    <span style="color: var(--text-muted); margin-left: 1rem;">{{.ImageCount}} images</span>
                </li>
                {{end}}
                {{end}}
            </ul>
        </div>
        {{end}}
    </div>

    <div class="footer">
        Generated by ecr-auditor
    </div>

    <script>
        function toggleImages(el) {
            el.classList.toggle('open');
            const imagesList = el.parentElement.querySelector('.images-list');
            imagesList.classList.toggle('open');
            if (imagesList.classList.contains('open')) {
                sortImagesTableByPushed(imagesList.querySelector('.images-table'));
            }
        }

        function filterTable() {
            const search = document.getElementById('search').value.toLowerCase();
            const policyFilter = document.getElementById('policyFilter').value;
            const rows = document.querySelectorAll('#repoTable > tbody > tr');

            rows.forEach(row => {
                const name = row.querySelector('td').textContent.toLowerCase();
                const hasPolicy = row.dataset.hasPolicy === 'true';

                let show = name.includes(search);
                if (policyFilter === 'with') show = show && hasPolicy;
                if (policyFilter === 'without') show = show && !hasPolicy;

                row.style.display = show ? '' : 'none';
            });
        }

        function sortTable(e, th, column) {
            e.stopPropagation();
            const table = document.getElementById('repoTable');
            const tbody = table.querySelector(':scope > tbody');
            const rows = Array.from(tbody.querySelectorAll(':scope > tr'));
            const headers = table.querySelectorAll('thead th');

            const currentlyAsc = th.classList.contains('sort-asc');
            const newDirection = currentlyAsc ? 'desc' : 'asc';
            headers.forEach(h => h.classList.remove('sort-asc', 'sort-desc'));
            th.classList.add('sort-' + newDirection);

            rows.sort((a, b) => {
                let aVal = a.dataset[column];
                let bVal = b.dataset[column];

                if (column === 'name') {
                    aVal = (aVal || '').toLowerCase();
                    bVal = (bVal || '').toLowerCase();
                } else {
                    aVal = parseFloat(aVal) || 0;
                    bVal = parseFloat(bVal) || 0;
                }

                let result = 0;
                if (aVal < bVal) result = -1;
                else if (aVal > bVal) result = 1;

                return newDirection === 'desc' ? -result : result;
            });

            rows.forEach(row => tbody.appendChild(row));
        }

        function sortImagesTableByPushed(table) {
            if (!table) return;
            const tbody = table.querySelector('tbody');
            const rows = Array.from(tbody.querySelectorAll('tr'));

            rows.sort((a, b) => {
                const aVal = parseInt(a.dataset.pushed);
                const bVal = parseInt(b.dataset.pushed);
                return bVal - aVal;
            });

            rows.forEach(row => tbody.appendChild(row));
        }

        function sortImagesTable(e, th, column) {
            e.stopPropagation();
            const table = th.closest('table');
            const tbody = table.querySelector('tbody');
            const rows = Array.from(tbody.querySelectorAll('tr'));
            const headers = table.querySelectorAll('thead th');

            const currentlyAsc = th.classList.contains('sort-asc');
            const newDirection = currentlyAsc ? 'desc' : 'asc';
            headers.forEach(h => h.classList.remove('sort-asc', 'sort-desc'));
            th.classList.add('sort-' + newDirection);

            rows.sort((a, b) => {
                let aVal, bVal;

                if (column === 'tags') {
                    aVal = (a.dataset[column] || '').toLowerCase();
                    bVal = (b.dataset[column] || '').toLowerCase();
                } else {
                    aVal = parseInt(a.dataset[column]) || 0;
                    bVal = parseInt(b.dataset[column]) || 0;
                }

                let result = 0;
                if (aVal < bVal) result = -1;
                else if (aVal > bVal) result = 1;

                return newDirection === 'desc' ? -result : result;
            });

            rows.forEach(row => tbody.appendChild(row));
        }

        function toggleTheme() {
            const html = document.documentElement;
            const currentTheme = html.getAttribute('data-theme');
            const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
            html.setAttribute('data-theme', newTheme);
            localStorage.setItem('theme', newTheme);
            updateThemeButton(newTheme);
        }

        function updateThemeButton(theme) {
            const btn = document.querySelector('.theme-toggle');
            btn.textContent = theme === 'dark' ? 'Toggle Light Mode' : 'Toggle Dark Mode';
        }

        (function() {
            const savedTheme = localStorage.getItem('theme');
            if (savedTheme) {
                document.documentElement.setAttribute('data-theme', savedTheme);
                updateThemeButton(savedTheme);
            } else if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
                document.documentElement.setAttribute('data-theme', 'dark');
                updateThemeButton('dark');
            }
        })();
    </script>
</body>
</html>`

type ReportGenerator struct {
	tmpl *template.Template
}

func NewReportGenerator() (*ReportGenerator, error) {
	funcMap := template.FuncMap{
		"formatBytes": formatBytes,
	}

	tmpl, err := template.New("report").Funcs(funcMap).Parse(reportTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse report template: %w", err)
	}

	return &ReportGenerator{tmpl: tmpl}, nil
}

func (g *ReportGenerator) Generate(w io.Writer, result *AuditResult) error {
	return g.tmpl.Execute(w, result)
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

type ReportData struct {
	*AuditResult
	GeneratedAt time.Time
}
