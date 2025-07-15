-- Valores totais arredondados para 2 casas decimais por condition do customer

SELECT 
    `condition`,
    ROUND(SUM(i.total), 2) AS total_rounded
FROM 
    customers c
LEFT JOIN 
    invoices i ON c.id = i.customer_id
GROUP BY 
    `condition`

-- Top 5 dos products mais vendidos e suas quantidades vendidas

SELECT 
    p.id,
    p.description,
    SUM(s.quantity) AS total_quantity_sold
FROM 
    products p
JOIN 
    sales s ON p.id = s.product_id
GROUP BY 
    p.id, p.description
ORDER BY 
    total_quantity_sold DESC
LIMIT 5;

-- Top 5 dos customers ativos quem gastou mais dinheiro                

SELECT 
    c.id,
    c.first_name,
    c.last_name,
    ROUND(SUM(i.total), 2) AS total_spent
FROM 
    customers c
JOIN 
    invoices i ON c.id = i.customer_id
WHERE 
    c.`condition` = 1
GROUP BY 
    c.id, c.first_name, c.last_name
ORDER BY 
    total_spent DESC
LIMIT 5;

