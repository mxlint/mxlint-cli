
document.filter = ["HIGH", "MEDIUM", "LOW"];
const filtersAppliedKey = 'filtersApplied';
function postMessage(message, data) {
    if (window.chrome.webview === undefined) {
        console.log("Missing webview ", message, data);
        return;
    }
    console.log("PostMessage", message, data);
    window.chrome.webview.postMessage({ message, data });
}

async function handleMessage(event) {
    console.log(event);
    const { message, data } = event.data;
    if (message === "refreshData") {
        await refreshData();
    } else if (message === "start") {
        document.getElementById("loading").classList.remove("hidden");
        document.getElementById("ready").classList.add("hidden");
    } else if (message === "end") {
        document.getElementById("loading").classList.add("hidden");
        document.getElementById("ready").classList.remove("hidden");
    }
}

if (window.chrome.webview !== undefined) {
    window.chrome.webview.addEventListener("message", handleMessage);
}


function getRule(path, rules) {
    for (const rule of rules) {
        if (rule.path == path) {
            return rule;
        }
    }
}

function flattenTestCase(testsuite, testcase, rules) {
    const rule = getRule(testsuite.name, rules);
    let status = "pass";
    let statusClass = "pico-background-cyan";
    if (rule.skipReason != "") {
        status = "skip";
        statusClass = "pico-background-slate";
    }
    if (testcase.failure) {
        status = "fail";
        statusClass = "pico-background-orange";
    }
    testcase.rule = rule;
    testcase.status = status;
    testcase.statusClass = statusClass;
    // clean up name
    let modelsource = "modelsource\\";
    if (testcase.name.startsWith(modelsource)) {
        testcase.name = testcase.name.substring(modelsource.length);
    }
    return testcase;
}

function createSpan(text, className) {
    let span = document.createElement("span");
    span.innerText = text;
    if (className !== undefined) {
        span.classList.add(className);
    }
    return span;
}

function createLink(text, href, obj) {
    let a = document.createElement("a");
    a.innerText = text;
    a.href = href;
    if (obj !== undefined) {
        a.addEventListener('click', (e) => {
            postMessage("openDocument", {
                document: obj.docname,
                type: obj.doctype,
                module: obj.module
            });
            e.preventDefault();
        });
    }
    return a;
}

function renderTestCase(testcase) {
    let tr = document.createElement("tr");
    let tdSeverity = document.createElement("td");
    tdSeverity.setAttribute("data-label", "Severity");
    let tdDocument = document.createElement("td");
    tdDocument.setAttribute("data-label", "Document");
    let tdModule = document.createElement("td");
    tdModule.setAttribute("data-label", "Module");
    let tdDocType = document.createElement("td");
    tdDocType.setAttribute("data-label", "Type");
    let tdRuleName = document.createElement("td");
    tdRuleName.setAttribute("data-label", "Rule");
    let tdCategory = document.createElement("td");
    tdCategory.setAttribute("data-label", "Category");
    let tdStatus = document.createElement("td");
    tdStatus.setAttribute("data-label", "Status");

    let details = document.createElement("details");
    let summary = document.createElement("summary");
    summary.innerText = testcase.rule.ruleName;
    details.appendChild(summary);

    let pDescription = document.createElement("p");
    let title = document.createElement("strong");
    title.innerText = testcase.rule.title;
    let description = document.createElement("span");
    description.innerText = testcase.rule.description;


    pDescription.appendChild(title);
    pDescription.appendChild(document.createElement("br"));
    pDescription.appendChild(description);
    details.appendChild(pDescription);

    let pRemediation = document.createElement("p");
    let remediation = document.createElement("strong");
    remediation.innerText = "Remediation";
    let remediationDescription = document.createElement("span");
    remediationDescription.innerText = testcase.rule.remediation;
    pRemediation.appendChild(remediation);
    pRemediation.appendChild(document.createElement("br"));
    pRemediation.appendChild(remediationDescription);
    pRemediation.classList.add("pico-color-blue");
    details.appendChild(pRemediation);

    if (testcase.status === "fail") {
        let pError = document.createElement("p");
        let error = document.createElement("strong");
        error.innerText = "Error";
        let errorDescription = document.createElement("span");
        errorDescription.innerText = testcase.failure.message;
        pError.appendChild(error);
        pError.appendChild(document.createElement("br"));
        pError.appendChild(errorDescription);
        pError.classList.add("pico-color-orange");
        details.appendChild(pError);
    }


    let spanStatus = document.createElement("span");
    spanStatus.innerText = testcase.status;
    spanStatus.classList.add("label");
    spanStatus.classList.add(testcase.statusClass);
    spanStatus.addEventListener('click', () => {
        postMessage("openDocument", { document: testcase.name });
    });

    tdSeverity.replaceChildren(createSpan(testcase.rule.severity));

    if (testcase.docname === "Metadata" && testcase.doctype === "") {
        tdDocument.replaceChildren(createSpan(testcase.docname));
    } else if (testcase.docname === "Security$ProjectSecurity" && testcase.doctype === "") {
        tdDocument.replaceChildren(createSpan(testcase.docname));
    } else {
        tdDocument.replaceChildren(createLink(testcase.docname, "#", testcase));
    }

    tdRuleName.replaceChildren(details);
    tdCategory.replaceChildren(createSpan(testcase.rule.category));
    tdDocType.replaceChildren(createSpan(testcase.doctype));
    tdModule.replaceChildren(createSpan(testcase.module));
    tdStatus.replaceChildren(spanStatus);

    tr.appendChild(tdSeverity);
    tr.appendChild(tdDocument);
    tr.appendChild(tdModule);
    tr.appendChild(tdDocType);
    tr.appendChild(tdRuleName);
    tr.appendChild(tdCategory);
    tr.appendChild(tdStatus);
    return tr;
}

function allowTestCaseRender(testCase) {
    const filterContent = document.getElementById('filter-content');
    let filtersApplied = filterContent.getAttribute(filtersAppliedKey);
    if (!filtersApplied) {
        return true;
    }

    let allow = true;

    //check severity
    filtersApplied = JSON.parse(filtersApplied);
    let severityFilters = filtersApplied['Severity'];
    //if (severityFiltersApplied && severityFiltersApplied != []) {
    //    allow = severityFiltersApplied.includes(testCase.rule.severity);
    //}
    allow = severityFilters.length === 0 || severityFilters.includes(testCase.rule.severity);

    return allow;
}

function renderData() {
    let details = document.getElementById("testcases");

    let ruleItems = [];
    let pass = 0;
    let skip = 0;
    let fail = 0;
    let total = 0;
    let all_testcases = [];
    let data = document.data;

    for (const testsuite of data.testsuites) {
        let testcases = testsuite.testcases;
        for (const testcase of testcases) {
            let ts = flattenTestCase(testsuite, testcase, data.rules);
            if (allowTestCaseRender(ts)) {
                if (ts.status === "fail") {
                    fail++;
                    ts.status_code = 1;
                } else if (ts.status === "skip") {
                    skip++;
                    ts.status_code = 2;
                } else {
                    pass++;
                    ts.status_code = 3;
                }
                if (ts.rule.severity === "HIGH") {
                    ts.severity_code = 1;
                } else if (ts.rule.severity === "MEDIUM") {
                    ts.severity_code = 2;
                } else {
                    ts.severity_code = 3;
                }
                const tokens = ts.name.split("\\");
                //console.log(tokens);
                ts.module = "";
                if (tokens.length > 1) {
                    ts.module = tokens[0];
                    const last = tokens.length - 1;
                    const rest = tokens.slice(1, tokens.length);
                    //console.log(rest);
                    if (rest.length > 1) {
                        ts.docname = rest.join("/").split('.')[0];
                        ts.doctype = tokens[last].split('.')[1];
                    } else {
                        ts.docname = tokens[last].split('.')[0]
                        ts.doctype = "";
                    }
                } else {
                    ts.docname = ts.name.split('.')[0];
                    ts.doctype = "";

                }
                all_testcases.push(ts);
            }
        }
    }

    let testcases_filtered = all_testcases.filter((ts) => document.filter.includes(ts.rule.severity));

    let testcases_sorted = testcases_filtered.sort((a, b) => {
        return a.status_code - b.status_code || a.severity_code - b.severity_code;
    });

    for (const ts of testcases_sorted) {
        let tr = renderTestCase(ts);
        ruleItems.push(tr);
    }
    let rules = data.rules.length;

    total = pass + skip + fail;
    let passWidth = (pass / total) * 100;
    let skipWidth = (skip / total) * 100;
    let failWidth = (fail / total) * 100;
    document.getElementById("summaryPass").style = "width: " + passWidth + "%;";
    document.getElementById("summarySkip").style = "width: " + skipWidth + "%;";
    document.getElementById("summaryFail").style = "width: " + failWidth + "%;";

    document.getElementById("pass").innerText = pass;
    document.getElementById("skip").innerText = skip;
    document.getElementById("fail").innerText = fail;
    document.getElementById("total").innerText = total;
    document.getElementById("rules").innerText = rules;


    if (total === 0) {
        console.log("No testcases found");
    }
    //else {
        details.replaceChildren(...ruleItems);
    //}
}

function applyFilter(e) {
    const filterContent = document.getElementById('filter-content');
    const filterType = e.getAttribute('parentFilter');

    // Initialize or retrieve filtersApplied
    let filtersApplied = filterContent.hasAttribute(filtersAppliedKey) ? JSON.parse(filterContent.getAttribute(filtersAppliedKey)) : {};

    // Ensure the filterType array exists
    filtersApplied[filterType] = filtersApplied[filterType] || [];

    // Add or remove the filter value based on checkbox status
    e.checked
        ? filtersApplied[filterType].push(e.value)
        : filtersApplied[filterType] = filtersApplied[filterType].filter(item => item !== e.value)

    //let filterHash = djb2(JSON.stringify(filtersApplied));
    //filterContent['hash'] = filterHash;

    // Update the attribute and notify
    filterContent.setAttribute(filtersAppliedKey, JSON.stringify(filtersApplied));
    console.log("Filters changed");
    renderData();
}

function djb2(str) {
    let hash = 5381;
    for (let i = 0; i < str.length; i++) {
        hash = (hash * 33) ^ str.charCodeAt(i);
    }
    return hash;
}

async function refreshData() {
    let response;
    if (window.chrome.webview === undefined) {
        response = await fetch("./api-sample.json");
    } else {
        response = await fetch("./api");
    }
    document.data = await response.json();
    let text = JSON.stringify(document.data);
    const newHash = djb2(text);
    if (document.hash !== newHash) {
        console.log("Data changed");
        renderData();
    }
    document.hash = newHash;
}

function init() {
    document.hash = "";
    document.data = {
        "testsuites": [],
        "rules": []
    }
    if (window.chrome.webview === undefined) {
        refreshData();
    }
    renderData();
}


document.getElementById("toggleDebug").addEventListener("click", () => {
    let hidden = document.getElementById("debug").classList.contains("hidden");
    if (hidden) {
        document.getElementById("debug").classList.remove("hidden");
    } else {
        document.getElementById("debug").classList.add("hidden");
    }
    postMessage("toggeDebug");

});

document.getElementById("btn-filter").addEventListener("click", () => {
    let filterContent = document.getElementById("filter-content")
    let hidden = filterContent.classList.contains("hidden");
    if (hidden) {
        filterContent.classList.remove("hidden");
    } else {
        filterContent.classList.add("hidden");
    }
});

document.addEventListener("DOMContentLoaded", () => {
    const checkboxes = document.querySelectorAll(".filter-checkbox");
    checkboxes.forEach((checkbox) => {
        checkbox.addEventListener("change", (event) => applyFilter(event.target));
    });
});

init();

postMessage("MessageListenerRegistered");
setInterval(async () => {
    postMessage("refreshData");

    await refreshData();
}, 1000);

