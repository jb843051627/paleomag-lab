async function load() {
  const specimens = await fetch('/api/specimens?limit=20').then((r) => r.json());
  document.querySelector('#specimens').innerHTML = (specimens.data || []).map((item) => `<div class="item"><span>${item.code}</span><small>${item.status}</small></div>`).join('') || '<p>暂无试样</p>';
  document.querySelector('#reviews').innerHTML = '<p>复核队列通过解释接口读取。</p>';
}
