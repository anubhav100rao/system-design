import time
import mysql.connector

cfg = {
    'user': 'root', 'password': 'dbpass',
    'host': 'localhost', 'database': 'system_design',
    'port': 3306
}


def measure(query, runs=20):
    conn = mysql.connector.connect(**cfg)
    cur = conn.cursor()
    # Warm-up
    cur.execute(query)
    _ = cur.fetchall()
    # Measure
    total = 0.0
    for _ in range(runs):
        start = time.time()
        cur.execute(query)
        _ = cur.fetchall()
        total += time.time() - start
    cur.close()
    conn.close()
    print(f"{query[:40]}... avg time: {total/runs:.6f} sec over {runs} runs")


if __name__ == "__main__":
    q1 = "SELECT * FROM employees_noidx WHERE dept='Sales';"
    q2 = "SELECT * FROM employees_idx WHERE dept='Sales';"
    measure(q1)
    measure(q2)
    # For rare value:
    # ensure 'RareDept' row exists in both tables, then:
    q3 = "SELECT * FROM employees_noidx WHERE dept='RareDept';"
    q4 = "SELECT * FROM employees_idx WHERE dept='RareDept';"
    measure(q3)
    measure(q4)
