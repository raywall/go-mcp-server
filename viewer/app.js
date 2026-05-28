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
      }
    } catch (e) {
      console.error("Erro ao ler LocalStorage", e);
    }
  }
});

function clearLocalDB() {
  if (confirm("Tem certeza que deseja limpar a memória local?")) {
    localStorage.removeItem('rulesCatalog');
    location.reload();
  }
}

// ==========================================
// 2. LEITURA DOS ARQUIVOS E BUSCA
// ==========================================
document.getElementById('folderInput').addEventListener('change', async function (event) {
  const files = event.target.files;
  const parsedRules = [];

  for (let file of files) {
    if (file.name.endsWith('.yaml') || file.name.endsWith('.yml')) {
      const text = await file.text();
      try {
        const doc = jsyaml.load(text);
        if (doc && doc.rule_id) parsedRules.push(doc);
      } catch (e) { console.error(`Erro ao parsear ${file.name}:`, e); }
    }
  }

  if (parsedRules.length > 0) {
    try { localStorage.setItem('rulesCatalog', JSON.stringify(parsedRules)); } catch (e) { }
    processAndRenderRules(parsedRules);
  } else {
    alert("Nenhum arquivo YAML válido contendo regras foi encontrado.");
  }
});

document.getElementById('searchInput').addEventListener('input', function (e) {
  const term = e.target.value.toLowerCase();
  const items = document.querySelectorAll('.rule-item');

  // 1. Filtra os itens individuais
  items.forEach(item => {
    const text = item.textContent.toLowerCase();
    if (text.includes(term)) {
      item.style.display = 'flex';
      if (term !== "") {
        const domainGroup = item.closest('.domain-group');
        const contextGroup = item.closest('.context-group');
        if (domainGroup) domainGroup.classList.remove('collapsed');
        if (contextGroup) contextGroup.classList.remove('collapsed');
      }
    } else {
      item.style.display = 'none';
    }
  });

  // 2. Esconde grupos de Contexto que ficaram vazios
  document.querySelectorAll('.context-group').forEach(group => {
    // Pega todos os itens visíveis dentro deste grupo
    const visibleItems = Array.from(group.querySelectorAll('.rule-item')).filter(i => i.style.display !== 'none');
    group.style.display = visibleItems.length === 0 && term !== "" ? 'none' : 'block';
  });

  // 3. Esconde grupos de Domínio que ficaram vazios
  document.querySelectorAll('.domain-group').forEach(group => {
    const visibleItems = Array.from(group.querySelectorAll('.rule-item')).filter(i => i.style.display !== 'none');
    group.style.display = visibleItems.length === 0 && term !== "" ? 'none' : 'block';
  });
});

// ==========================================
// 3. RENDERIZAÇÃO E ACCORDION
// ==========================================
function processAndRenderRules(rules) {
  rules.sort((a, b) => {
    if (a.domain !== b.domain) return a.domain.localeCompare(b.domain);
    if (a.context !== b.context) return (a.context || "").localeCompare(b.context || "");
    return (a.execution_order || 0) - (b.execution_order || 0);
  });

  allRulesFlat = rules;
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

    // onclick injetado para dar toggle na classe .collapsed do parentElement
    domainDiv.innerHTML = `
            <div class="domain-title" onclick="this.parentElement.classList.toggle('collapsed')">
                <span>${domain}</span> 
                <i class="fas fa-chevron-down chevron"></i>
            </div>
            <div class="item-container domain-container"></div>
        `;

    const dContainer = domainDiv.querySelector('.domain-container');

    for (const context in tree[domain]) {
      const contextDiv = document.createElement('div');
      contextDiv.className = 'context-group';

      let cContainer = contextDiv; // Por padrão as regras entram direto aqui

      if (context !== "Geral") {
        contextDiv.innerHTML = `
                    <div class="context-title" onclick="this.parentElement.classList.toggle('collapsed')">
                        <span>${context}</span>
                        <i class="fas fa-chevron-down chevron"></i>
                    </div>
                    <div class="item-container context-container"></div>
                `;
        cContainer = contextDiv.querySelector('.context-container');
      }

      tree[domain][context].forEach(rule => {
        const ruleDiv = document.createElement('div');
        ruleDiv.className = 'rule-item';
        ruleDiv.id = `menu_item_${rule._flatIndex}`;

        const name = rule.human_context?.name || "Regra sem nome";
        const owner = rule.human_context?.business_owner || "Sem Owner";

        // Texto injetado dentro da tag escondida com ID para a busca varrer corretamente
        ruleDiv.innerHTML = `
                    <div style="display:none">${rule.rule_id} ${rule.context}</div>
                    
                    <div class="rule-badges">
                        <span class="badge-context">${context}</span>
                        <span class="order-badge" title="Ordem de Execução"><i class="fas fa-layer-group"></i> ${rule.execution_order || '-'}</span>
                    </div>
                    
                    <div class="rule-name">${name}</div>
                    <div class="rule-id">${rule.rule_id}</div>
                    
                    <div class="rule-meta">
                        <span class="meta-label"><i class="fas fa-user-tie"></i> Owner:</span>
                        <span class="badge-owner">${owner}</span>
                    </div>
                `;

        ruleDiv.onclick = () => loadRule(rule._flatIndex);
        cContainer.appendChild(ruleDiv);
      });
      dContainer.appendChild(contextDiv);
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
// 5. REDIMENSIONAMENTO (RESIZER)
// ==========================================
const resizer = document.getElementById('dragMe');
const sidebar = document.getElementById('sidebar');
let x = 0;
let w = 0;

const mouseMoveHandler = function (e) {
  const dx = e.clientX - x;
  sidebar.style.width = `${w + dx}px`;
};

const mouseUpHandler = function () {
  document.removeEventListener('mousemove', mouseMoveHandler);
  document.removeEventListener('mouseup', mouseUpHandler);
  localStorage.setItem('sidebarWidth', sidebar.style.width);
};

resizer.addEventListener('mousedown', function (e) {
  x = e.clientX;
  w = sidebar.getBoundingClientRect().width;
  document.addEventListener('mousemove', mouseMoveHandler);
  document.addEventListener('mouseup', mouseUpHandler);
});

document.addEventListener('DOMContentLoaded', () => {
  const savedWidth = localStorage.getItem('sidebarWidth');
  if (savedWidth) sidebar.style.width = savedWidth;
});