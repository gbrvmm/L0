const $ = (s) => document.querySelector(s);
const statusEl = $("#status");
const resultEl = $("#result");
const resIdEl = $("#res-id");
const resDeliveryEl = $("#res-delivery");
const resPaymentEl = $("#res-payment");
const itemsBody = $("#items tbody");

function fmtMoney(n) {
  return new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'USD' }).format(n);
}

function createInfoTable(data) {
  let html = '<table class="info-table"><tbody>';
  for (const [key, value] of Object.entries(data)) {
    const displayKey = key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
    html += `
      <tr>
        <th>${displayKey}</th>
        <td>${value !== null && value !== undefined ? value : '-'}</td>
      </tr>
    `;
  }
  html += '</tbody></table>';
  return html;
}

function showOrder(o) {
  resultEl.classList.remove("hidden");
  resIdEl.textContent = o.order_uid;
  resDeliveryEl.innerHTML = createInfoTable(o.delivery);
  resPaymentEl.innerHTML = createInfoTable(o.payment);

  itemsBody.innerHTML = "";
  (o.items || []).forEach(it => {
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td>${it.name}</td>
      <td>${it.brand}</td>
      <td>${fmtMoney(it.price)}</td>
      <td>${it.sale}%</td>
      <td>${fmtMoney(it.total_price)}</td>
      <td><code>${it.rid}</code></td>
    `;
    itemsBody.appendChild(tr);
  });
}

async function fetchOrder(id) {
  statusEl.textContent = "Загрузка…";
  resultEl.classList.add("hidden");
  try {
    const res = await fetch(`/api/orders/${encodeURIComponent(id)}`);
    if (res.ok) {
      const data = await res.json();
      showOrder(data);
      statusEl.textContent = "Готово";
    } else if (res.status === 404) {
      statusEl.textContent = "Не найдено";
    } else {
      statusEl.textContent = "Ошибка: " + res.status;
    }
  } catch (e) {
    console.error(e);
    statusEl.textContent = "Сетевая ошибка";
  }
}

$("#findBtn").addEventListener("click", () => {
  const id = $("#orderId").value.trim();
  if (!id) return;
  fetchOrder(id);
});

$("#sampleBtn").addEventListener("click", () => {
  $("#orderId").value = "b563feb7b2b84b6test";
});

$("#orderId").addEventListener("keydown", (e) => {
  if (e.key === "Enter") $("#findBtn").click();
});