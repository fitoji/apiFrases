import json
import os
import sys
import urllib.request
import urllib.error

url = os.environ.get("API_URL") or (sys.argv[1] if len(sys.argv) > 1 else "http://localhost:8000/frases")
headers = {"Content-Type": "application/json"}

data = [
  {
    "id": 1,
    "frase": "Somos lo que pensamos. Todo lo que somos surge con nuestros pensamientos. Con nuestros pensamientos construimos el mundo.",
    "original": "Manopubbaṅgamā dhammā, manoseṭṭhā manomayā",
    "autor": "Buda",
    "categoria": "budismo"
  },
  {
    "id": 2,
    "frase": "No hay camino hacia la felicidad, la felicidad es el camino.",
    "original": "Natthi santaraṁ sukhaṁ",
    "autor": "Buda",
    "categoria": "budismo"
  },
  {
    "id": 3,
    "frase": "El dolor es inevitable, el sufrimiento es opcional.",
    "original": "Sabbe saṅkhārā aniccā, sabbe saṅkhārā dukkhā",
    "autor": "Buda",
    "categoria": "budismo"
  },
  {
    "id": 4,
    "frase": "La paz viene de dentro. No la busques fuera.",
    "original": "Ajjhattaṁ yeva pariyesati sukhaṁ",
    "autor": "Buda",
    "categoria": "budismo"
  },
  {
    "id": 5,
    "frase": "No creas en nada simplemente porque lo has oído. Examínalo por ti mismo.",
    "original": "Mā anussavena, attanāva jānāhi",
    "autor": "Buda",
    "categoria": "budismo"
  },
  {
    "id": 6,
    "frase": "El odio nunca se extingue con el odio, sino con el amor. Esta es una ley eterna.",
    "original": "Na hi verena verāni sammantīdha kudācanaṁ, averena ca sammanti esa dhammo sanantano",
    "autor": "Buda",
    "categoria": "budismo"
  },
  {
    "id": 7,
    "frase": "La salud es el mayor regalo, la satisfacción la mayor riqueza, la fidelidad la mejor relación.",
    "original": "Ārogyaparamā lābhā, santuṭṭhiparamaṁ dhanaṁ, vissāsaparamā ñāti",
    "autor": "Buda",
    "categoria": "budismo"
  },
  {
    "id": 8,
    "frase": "Tres cosas no pueden permanecer ocultas: el sol, la luna y la verdad.",
    "original": "Tisso imā aparihīnā nappahīyanti kudācanaṁ",
    "autor": "Buda",
    "categoria": "budismo"
  },
  {
    "id": 9,
    "frase": "Trabaja tu propia salvación. No dependas de otros.",
    "original": "Attahi attano nātho, ko hi nātho paro siyā",
    "autor": "Buda",
    "categoria": "budismo"
  },
  {
    "id": 10,
    "frase": "El que está libre de pensamientos perturbadores disfruta de la paz perfecta.",
    "original": "Akkodhena jine kodhaṁ, asādhuṁ sādhunā jine",
    "autor": "Buda",
    "categoria": "budismo"
  }
]

print(f"🚀 Iniciando carga de {len(data)} frases a {url}...")

for item in data:
    try:
        # Convertimos el diccionario a JSON bytes
        json_data = json.dumps(item).encode('utf-8')
        
        # Creamos la solicitud POST
        req = urllib.request.Request(url, data=json_data, headers=headers, method='POST')
        
        # Enviamos la solicitud
        with urllib.request.urlopen(req) as response:
            print(f"✅ Frase {item['id']} insertada: {response.status}")
            
    except urllib.error.HTTPError as e:
        print(f"❌ Error insertando frase {item['id']}: {e.code} - {e.reason}")
    except Exception as e:
        print(f"❌ Error de conexión con frase {item['id']}: {e}")

print("🏁 Proceso finalizado.")