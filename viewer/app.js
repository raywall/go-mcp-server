let allRulesFlat = [];
let currentIndex = -1;

// ==========================================
// 1. INICIALIZAÇÃO E LOCAL STORAGE
// ==========================================
document.addEventListener('DOMContentLoaded', () => {
  const savedRules = localStorage.getItem('rulesCatalog');
  if (savedRules) {
    try {
      const rules = JSON.parse(savedRules);
      if (rules.length > 0) {
        processAndRenderRules(rules);
        console.log("Regras carregadas da memória local.");
      }
    } catch (e) {
      console.error("Erro ao ler LocalStorage", e);
    }
  }
});

function clearLocalDB() {
  localStorage.removeItem('rulesCatalog');
  location.reload();
}

// ==========================================
// 2. LEITURA DOS ARQUIVOS LOCAIS
// ==========================================
document.getElementById('folderInput').addEventListener('change', async function (event) {
  const files = event.target.files;
  const parsedRules = [];

  for (let file of files) {
    if (file.name.endsWith('.yaml') || file.name.endsWith('.yml')) {
      const text = await file.text();
      try {
        const doc = jsyaml.load(text);
        if (doc && doc.rule_id) {
          parsedRules.push(doc);
        }
      } catch (e) {
        console.error(`Erro ao parsear ${file.name}:`, e);
      }
    }
  }

  if (parsedRules.length > 0) {
    // Salva no LocalDB para não perder no refresh
    try {
      localStorage.setItem('rulesCatalog', JSON.stringify(parsedRules));
    } catch (e) {
      alert("Aviso: Muitas regras carregadas. O limite do navegador foi atingido e elas não serão salvas no refresh.");
    }
    processAndRenderRules(parsedRules);
  } else {
    alert("Nenhum arquivo YAML válido contendo regras foi encontrado.");
  }
});

// ==========================================
// 3. RENDERIZAÇÃO DA INTERFACE (NOVO LAYOUT)
// ==========================================
function processAndRenderRules(rules) {
  // Ordenação: Domínio -> Contexto -> Ordem
  rules.sort((a, b) => {
    if (a.domain !== b.domain) return a.domain.localeCompare(b.domain);
    if (a.context !== b.context) return (a.context || "").localeCompare(b.context || "");
    return (a.execution_order || 0) - (b.execution_order || 0);
  });

  allRulesFlat = rules;

  // Agrupamento
  const tree = {};
  rules.forEach((rule, index) => {
    rule._flatIndex = index;
    const dom = rule.domain || "Geral";
    const ctx = rule.context || "Geral";
    if (!tree[dom]) tree[dom] = {};
    if (!tree[dom][ctx]) tree[dom][ctx] = [];
    tree[dom][ctx].push(rule);
  });

  const listEl = document.getElementById('ruleList');
  listEl.innerHTML = '';

  for (const domain in tree) {
    const domainDiv = document.createElement('div');
    domainDiv.className = 'domain-group';
    domainDiv.innerHTML = `<div class="domain-title">${domain}</div>`;

    for (const context in tree[domain]) {
      const contextDiv = document.createElement('div');
      contextDiv.className = 'context-group';

      // Exibe o contexto apenas se não for "Geral" para evitar repetição
      if (context !== "Geral") {
        contextDiv.innerHTML = `<div class="context-title">${context}</div>`;
      }

      tree[domain][context].forEach(rule => {
        const ruleDiv = document.createElement('div');
        ruleDiv.className = 'rule-item';
        ruleDiv.id = `menu_item_${rule._flatIndex}`;

        // Novo Layout Rica do Card
        const name = rule.human_context?.name || "Regra sem nome";
        const owner = rule.human_context?.business_owner || "Sem Owner";

        ruleDiv.innerHTML = `
                    <div class="rule-name"><span class="order-badge">${rule.execution_order || '-'}</span> ${name}</div>
                    <div class="rule-id">${rule.rule_id}</div>
                    <div class="rule-meta">
                        <span class="badge-context">${context}</span>
                        <span class="badge-owner">${owner}</span>
                    </div>
                `;

        ruleDiv.onclick = () => loadRule(rule._flatIndex);
        contextDiv.appendChild(ruleDiv);
      });
      domainDiv.appendChild(contextDiv);
    }
    listEl.appendChild(domainDiv);
  }
}

// ==========================================
// 4. FORMULÁRIO E NAVEGAÇÃO
// ==========================================
function loadRule(index) {
  if (index < 0 || index >= allRulesFlat.length) return;
  currentIndex = index;
  const rule = allRulesFlat[index];

  document.querySelectorAll('.rule-item').forEach(el => el.classList.remove('active'));
  document.getElementById(`menu_item_${index}`).classList.add('active');

  document.getElementById('emptyState').style.display = 'none';
  document.getElementById('formContainer').style.display = 'block';

  const setVal = (id, val) => document.getElementById(id).value = val || '';

  document.getElementById('f_rule_id').innerText = rule.rule_id;
  document.getElementById('f_status').innerText = rule.status || 'ACTIVE';

  setVal('f_domain', rule.domain);
  setVal('f_context', rule.context);
  setVal('f_execution_order', rule.execution_order);
  setVal('f_owner', rule.human_context?.business_owner);
  setVal('f_name', rule.human_context?.name);
  setVal('f_description', rule.human_context?.description);
  setVal('f_app_name', rule.engineering_context?.application_name);
  setVal('f_app_type', rule.engineering_context?.application_type);
  setVal('f_repo', rule.engineering_context?.repository_url);
  setVal('f_entrypoint', rule.engineering_context?.entrypoint);
  setVal('f_ai_logic', rule.ai_logic);
  setVal('f_monitor', rule.technical_metadata?.observability?.datadog_monitor_id);
  setVal('f_log_marker', rule.technical_metadata?.observability?.log_marker);

  document.getElementById('btnPrev').disabled = (currentIndex === 0);
  document.getElementById('btnNext').disabled = (currentIndex === allRulesFlat.length - 1);
}

function navigate(direction) {
  loadRule(currentIndex + direction);
}

// ==========================================
// 5. LÓGICA DE REDIMENSIONAMENTO DA SIDEBAR
// ==========================================
const resizer = document.getElementById('dragMe');
const sidebar = document.getElementById('sidebar');

let x = 0;
let w = 0;

const mouseMoveHandler = function (e) {
  const dx = e.clientX - x;
  const newWidth = w + dx;
  // Respeita os limites definidos no CSS (min-width e max-width)
  sidebar.style.width = `${newWidth}px`;
};

const mouseUpHandler = function () {
  document.removeEventListener('mousemove', mouseMoveHandler);
  document.removeEventListener('mouseup', mouseUpHandler);
  // Opcional: Salvar a nova largura no LocalStorage também!
  localStorage.setItem('sidebarWidth', sidebar.style.width);
};

resizer.addEventListener('mousedown', function (e) {
  x = e.clientX;
  w = sidebar.getBoundingClientRect().width;

  document.addEventListener('mousemove', mouseMoveHandler);
  document.addEventListener('mouseup', mouseUpHandler);
});

// Restaura a largura salva, se houver
document.addEventListener('DOMContentLoaded', () => {
  const savedWidth = localStorage.getItem('sidebarWidth');
  if (savedWidth) {
    sidebar.style.width = savedWidth;
  }
});