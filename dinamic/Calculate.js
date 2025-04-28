function calculate() {
    const ip = document.getElementById("ip").value;
    const mask = document.getElementById("mask").value;

    if (!ip || !mask) {
        document.getElementById("result").innerHTML = `<p style="color:red;">Lütfen IP ve Mask değerlerini girin.</p>`;
        return;
    }

    fetch(`${document.location.protocol}//${document.location.host}/calculate?ip=${ip}&mask=${mask}`)
        .then(response => {
            if (!response.ok) {
                return response.text().then(err => { throw new Error(err); });
            }
            return response.json();
        })
        .then(data => {
            document.getElementById("result").innerHTML = `
                <div class="card p-3 mt-3 shadow">
                    <h3 class="text-center">${ip} IP Hesaplama</h3>
                    <p><strong>Network:</strong> ${data.network}</p>
                    <p><strong>Broadcast:</strong> ${data.broadcast}</p>
                    <p><strong>Host Min:</strong> ${data.host_min}</p>
                    <p><strong>Host Max:</strong> ${data.host_max}</p>
                </div>
            `;
        })
        .catch(error => {
            document.getElementById("result").innerHTML = `<p style="color:red;">Hata: ${error.message}</p>`;
        });
}

function calculatePost() {
    const ip = document.getElementById("ip").value;
    const mask = document.getElementById("mask").value;

    if (!ip || !mask) {
        document.getElementById("result").innerHTML = `<p style="color:red;">Lütfen IP ve Mask değerlerini girin.</p>`;
        return;
    }

    fetch(`${document.location.protocol}//${document.location.host}/calculatePost`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({ ip: ip, mask: mask })
    })
        .then(response => {
            if (!response.ok) {
                return response.text().then(err => { throw new Error(err); });
            }
            return response.json();
        })
        .then(data => {
            document.getElementById("result").innerHTML = `
                <div class="card p-3 mt-3 shadow">
                    <h3 class="text-center">${ip} IP Hesaplama</h3>
                    <p><strong>Network:</strong> ${data.network}</p>
                    <p><strong>Broadcast:</strong> ${data.broadcast}</p>
                    <p><strong>Host Min:</strong> ${data.host_min}</p>
                    <p><strong>Host Max:</strong> ${data.host_max}</p>
                </div>
            `;
        })
        .catch(error => {
            document.getElementById("result").innerHTML = `<p style="color:red;">Hata: ${error.message}</p>`;
        });
}