let allRulesFlat = []; // Array linear para navegação Prev/Next
let currentIndex = -1;

document.getElementById('folderInput').addEventListener('change', handleFileSelect);

async function handleFileSelect(event) {
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
        console.error(`Erro ao fazer parse do arquivo ${file.name}:`, e);
      }
    }
  }

  processAndRenderRules(parsedRules);
}

function processAndRenderRules(rules) {
  // 1. Organizar e ordenar os dados
  // Ordenação global: Domínio -> Contexto -> Ordem de Execução
  rules.sort((a, b) => {
    if (a.domain !== b.domain) return a.domain.localeCompare(b.domain);
    if (a.context !== b.context) return a.context.localeCompare(b.context);
    return a.execution_order - b.execution_order;
  });

  allRulesFlat = rules; // Salva o estado linear para navegação

  // 2. Agrupar para renderização do menu
  const tree = {};
  rules.forEach((rule, index) => {
    rule._flatIndex = index; // Guarda a posição original
    if (!tree[rule.domain]) tree[rule.domain] = {};
    if (!tree[rule.domain][rule.context]) tree[rule.domain][rule.context] = [];
    tree[rule.domain][rule.context].push(rule);
  });

  // 3. Renderizar Sidebar
  const listEl = document.getElementById('ruleList');
  listEl.innerHTML = '';

  for (const domain in tree) {
    const domainDiv = document.createElement('div');
    domainDiv.className = 'domain-group';
    domainDiv.innerHTML = `<div class="domain-title">${domain}</div>`;

    for (const context in tree[domain]) {
      const contextDiv = document.createElement('div');
      contextDiv.className = 'context-group';
      contextDiv.innerHTML = `<div class="context-title">${context}</div>`;

      tree[domain][context].forEach(rule => {
        const ruleDiv = document.createElement('div');
        ruleDiv.className = 'rule-item';
        ruleDiv.id = `menu_item_${rule._flatIndex}`;
        ruleDiv.innerHTML = `<span class="order-badge">${rule.execution_order}</span> ${rule.human_context?.name || rule.rule_id}`;
        ruleDiv.onclick = () => loadRule(rule._flatIndex);
        contextDiv.appendChild(ruleDiv);
      });
      domainDiv.appendChild(contextDiv);
    }
    listEl.appendChild(domainDiv);
  }
}

function loadRule(index) {
  if (index < 0 || index >= allRulesFlat.length) return;
  currentIndex = index;
  const rule = allRulesFlat[index];

  // Atualiza UI Visual Menu
  document.querySelectorAll('.rule-item').forEach(el => el.classList.remove('active'));
  document.getElementById(`menu_item_${index}`).classList.add('active');

  // Esconde estado vazio, mostra formulário
  document.getElementById('emptyState').style.display = 'none';
  document.getElementById('formContainer').style.display = 'block';

  // Popula os campos do form
  const setVal = (id, val) => document.getElementById(id).value = val || '';

  document.getElementById('f_rule_id').innerText = rule.rule_id;
  document.getElementById('f_status').innerText = rule.status;

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

  // Controla estado dos botões de navegação
  document.getElementById('btnPrev').disabled = (currentIndex === 0);
  document.getElementById('btnNext').disabled = (currentIndex === allRulesFlat.length - 1);
}

function navigate(direction) {
  loadRule(currentIndex + direction);
}