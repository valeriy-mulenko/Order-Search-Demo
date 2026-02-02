// Тестовый JSON данные
const TEST_JSON = {
    "order_id": "test1234567890",
    "client_id": 1234567890,
    "locale": "ru",
    "delivery": {
        "name": "Иван Иванов",
        "phone": "+71234567890",
        "email": "test@test.ru",
        "type": "PVZ",
        "city": "Saint-Petersburg",
        "address": "Turistskaya street, 10"
    },
    "payment": {
        "transaction_id": "payment_test4566435",
        "currency": "RUB",
        "provider": "OzonBank",
        "amount": 1791.00,
        "date_pay": 1756207484,
        "bank": "alpha"
    },
    "items": [
        {
            "product_id": 1136435021,
            "name": "T-shirt",
            "brand": "Ozon Russia",
            "price": 890.00,
            "size": "48",
            "quantity": 1
        },
        {
            "product_id": 1651699088,
            "name": "Grok the algorithms",
            "brand": "Peter Publishing House",
            "price": 901.00,
            "size": "",
            "quantity": 1
        }
    ],
    "date_created": "2025-08-26T14:24:44Z"
};

// Функция для показа ошибок
function showError(message, type = 'error') {
    const errorDiv = document.getElementById('error');
    if (!errorDiv) {
        console.error('Error div not found!');
        return;
    }
    errorDiv.textContent = message;
    errorDiv.className = 'error active';
    if (type === 'success') {
        errorDiv.style.backgroundColor = '#c6f6d5';
        errorDiv.style.color = '#22543d';
        errorDiv.style.borderLeft = '4px solid #38a169';
    }
}

// Функция для скрытия ошибок
function hideError() {
    const errorDiv = document.getElementById('error');
    if (errorDiv) {
        errorDiv.classList.remove('active');
    }
}

// Функция для показа/скрытия загрузки
function showLoading(show) {
    const loading = document.getElementById('loading');
    if (loading) {
        if (show) {
            loading.classList.add('active');
        } else {
            loading.classList.remove('active');
        }
    }
}

// Функция для показа/скрытия результата
function showResult() {
    const result = document.getElementById('result');
    if (result) {
        result.classList.add('active');
    }
}

function hideResult() {
    const result = document.getElementById('result');
    if (result) {
        result.classList.remove('active');
    }
}

// Основная функция поиска заказа
async function getOrder() {
    const orderIdInput = document.getElementById('orderId');
    if (!orderIdInput) {
        showError('Поле ввода не найдено');
        return;
    }

    const orderId = orderIdInput.value.trim();
    if (!orderId) {
        showError('Пожалуйста, введите ID заказа');
        return;
    }

    showLoading(true);
    hideError();
    hideResult();

    try {
        const response = await fetch(`/api/order?order_id=${encodeURIComponent(orderId)}`);

        if (response.status === 404) {
            showError('Заказ не найден');
            return;
        }

        if (!response.ok) {
            throw new Error(`Ошибка сервера: ${response.status}`);
        }

        const order = await response.json();
        displayOrder(order);
    } catch (error) {
        showError('Ошибка при получении заказа: ' + error.message);
        console.error('Get order error:', error);
    } finally {
        showLoading(false);
    }
}

// Функция создания заказа
async function createOrder() {
    const orderJsonInput = document.getElementById('orderJson');
    if (!orderJsonInput) {
        showError('Текстовое поле не найдено');
        return;
    }

    const orderJson = orderJsonInput.value.trim();
    if (!orderJson) {
        showError('Пожалуйста, введите данные заказа в формате JSON');
        return;
    }

    let orderData;
    try {
        orderData = JSON.parse(orderJson);
    } catch (error) {
        showError('Неверный формат JSON: ' + error.message);
        return;
    }

    showLoading(true);
    hideError();

    try {
        const response = await fetch('/api/order', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(orderData),
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(errorText || `Ошибка: ${response.status}`);
        }

        const result = await response.json();
        showError(`Заказ успешно создан! ID: ${result.order_id}`, 'success');

        // Очистка поля ввода
        orderJsonInput.value = '';

        // Вывод созданного заказа
        const orderIdInput = document.getElementById('orderId');
        if (orderIdInput) {
            orderIdInput.value = result.order_id;
            setTimeout(() => getOrder(), 500);
        }
    } catch (error) {
        showError('Ошибка при создании заказа: ' + error.message);
        console.error('Create order error:', error);
    } finally {
        showLoading(false);
    }
}

// Для копирования тестового JSON
function copyTestJson() {
    const jsonText = JSON.stringify(TEST_JSON, null, 2);
    const button = document.getElementById('copyJsonBtn');

    if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(jsonText)
            .then(() => {
                if (button) {
                    const originalText = button.textContent;
                    button.textContent = '✅ Скопировано!';
                    setTimeout(() => {
                        button.textContent = originalText;
                    }, 2000);
                }
            })
            .catch(err => {
                console.error('Copy error:', err);
                showError('Не удалось скопировать JSON');
            });
    } else {
        // Fallback для старых браузеров
        const textArea = document.createElement('textarea');
        textArea.value = jsonText;
        document.body.appendChild(textArea);
        textArea.select();
        try {
            document.execCommand('copy');
            if (button) {
                const originalText = button.textContent;
                button.textContent = '✅ Скопировано!';
                setTimeout(() => {
                    button.textContent = originalText;
                }, 2000);
            }
        } catch (err) {
            console.error('Fallback copy error:', err);
            showError('Не удалось скопировать JSON');
        } finally {
            document.body.removeChild(textArea);
        }
    }
}

// Функция отображения заказа
function displayOrder(order) {
    const orderDetails = document.getElementById('orderDetails');
    if (!orderDetails) {
        console.error('Order details container not found');
        return;
    }


    // Форматируем дату платежа
    const formatDate = (timestamp) => {
        try {
            const date = new Date(timestamp * 1000);
            return date.toLocaleString('ru-RU');
        } catch (e) {
            return timestamp;
        }
    };

    const html = `
        <div class="order-info">
            <div class="info-section">
                <h3>📦 Основная информация</h3>
                <div class="info-item">
                    <span class="info-label">ID заказа:</span>
                    <span class="info-value">${order.order_id}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">ID клиента:</span>
                    <span class="info-value">${order.client_id}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">Дата создания:</span>
                    <span class="info-value">${formatDate(order.payment.date_pay)}</span>
                </div>
            </div>
            
            <div class="info-section">
                <h3>🚚 Доставка</h3>
                <div class="info-item">
                    <span class="info-label">Получатель:</span>
                    <span class="info-value">${order.delivery.name}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">Телефон:</span>
                    <span class="info-value">${order.delivery.phone}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">Email:</span>
                    <span class="info-value">${order.delivery.email}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">Тип доставки:</span>
                    <span class="info-value">${order.delivery.type}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">Адрес:</span>
                    <span class="info-value">${order.delivery.city}, ${order.delivery.address}</span>
                </div>
            </div>
            
            <div class="info-section">
                <h3>💳 Оплата</h3>
                <div class="info-item">
                    <span class="info-label">Сумма:</span>
                    <span class="info-value">${order.payment.amount.toFixed(2)} ${order.payment.currency}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">Транзакция:</span>
                    <span class="info-value">${order.payment.transaction_id}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">Провайдер:</span>
                    <span class="info-value">${order.payment.provider}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">Банк:</span>
                    <span class="info-value">${order.payment.bank}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">Дата оплаты:</span>
                    <span class="info-value">${formatDate(order.payment.date_pay)}</span>
                </div>
            </div>
        </div>
        
        <div style="margin-top: 2rem;">
            <h3>🛍️ Товары (${order.items ? order.items.length : 0})</h3>
            ${order.items && order.items.length > 0 ? `
                <div style="overflow-x: auto;">
                    <table style="width: 100%; border-collapse: collapse; margin-top: 1rem;">
                        <thead>
                            <tr style="background: #667eea; color: white;">
                                <th style="padding: 0.75rem; text-align: left;">Название</th>
                                <th style="padding: 0.75rem; text-align: left;">Бренд</th>
                                <th style="padding: 0.75rem; text-align: left;">Цена</th>
                                <th style="padding: 0.75rem; text-align: left;">Количество</th>
                                <th style="padding: 0.75rem; text-align: left;">Размер</th>
                                <th style="padding: 0.75rem; text-align: left;">Итого</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${order.items.map(item => `
                                <tr style="border-bottom: 1px solid #eee;">
                                    <td style="padding: 0.75rem;">${item.name || ''}</td>
                                    <td style="padding: 0.75rem;">${item.brand || ''}</td>
                                    <td style="padding: 0.75rem;">${item.price ? item.price.toFixed(2) + ' ₽' : ''}</td>
                                    <td style="padding: 0.75rem;">${item.quantity || ''}</td>
                                    <td style="padding: 0.75rem;">${item.size || ''}</td>
                                    <td style="padding: 0.75rem;">${item.price && item.quantity ? (item.price * item.quantity).toFixed(2) + ' ₽' : ''}</td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                </div>
            ` : '<p>Товары не найдены</p>'}
        </div>
        
        <div style="margin-top: 2rem;">
            <h3>📄 Полные данные (JSON)</h3>
            <pre><code>${JSON.stringify(order, null, 2)}</code></pre>
        </div>
    `;

    orderDetails.innerHTML = html;
    showResult();
}

// Отображение тестовых заказов
function displayTestOrders(orders) {
    const testOrdersList = document.getElementById('testOrders');
    if (testOrdersList && orders && orders.length > 0) {
        testOrdersList.innerHTML = orders.map(order =>
            `<li><a href="#" onclick="document.getElementById('orderId').value='${order.order_id}'; getOrder(); return false;">${order.order_id}</a></li>`
        ).join('');
    }
}

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', function () {
    console.log('Page loaded, JavaScript is working!');

    // Загружаем тестовый заказ
    fetch('/api/order?order_id=test1234567890')
        .then(response => {
            if (response.ok) return response.json();
            throw new Error('Failed to fetch test order');
        })
        .then(order => displayTestOrders([order]))
        .catch(error => {
            console.log('No test orders found:', error.message);
            // Создаем тестовый элемент если API недоступно
            const testOrdersList = document.getElementById('testOrders');
            if (testOrdersList) {
                testOrdersList.innerHTML = '<li><a href="#" onclick="document.getElementById(\'orderId\').value=\'test1234567890\'; getOrder(); return false;">test1234567890</a></li>';
            }
        });

    // Обработчик Enter для поля поиска
    const orderIdInput = document.getElementById('orderId');
    if (orderIdInput) {
        orderIdInput.addEventListener('keypress', function (event) {
            if (event.key === 'Enter') {
                getOrder();
            }
        });
    }

    // Обработчик для кнопки копирования
    const copyButton = document.getElementById('copyJsonBtn');
    if (copyButton) {
        copyButton.addEventListener('click', copyTestJson);
    }
});