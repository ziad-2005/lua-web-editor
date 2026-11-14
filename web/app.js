document.getElementById('run-btn').addEventListener('click', async () => {
    const code = document.getElementById('code-editor').value;
    const runBtn = document.getElementById('run-btn');
    const statusBar = document.getElementById('status-bar');
    const stdoutEl = document.getElementById('stdout');
    const stderrEl = document.getElementById('stderr');
    const exitCodeEl = document.getElementById('exit-code');
    const durationEl = document.getElementById('duration');

    // Reset UI
    stdoutEl.textContent = '';
    stderrEl.textContent = '';
    exitCodeEl.textContent = '-';
    durationEl.textContent = '-';
    statusBar.textContent = 'Status: Running...';
    statusBar.style.background = '#d97706'; // amber
    runBtn.disabled = true;

    try {
        const response = await fetch('/api/runs', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ code })
        });

        if (!response.ok) {
            throw new Error(`Server returned status ${response.status}`);
        }

        const result = await response.json();

        stdoutEl.textContent = result.stdout;
        stderrEl.textContent = result.stderr;
        exitCodeEl.textContent = result.exit_code;
        durationEl.textContent = result.duration_ms;
        statusBar.textContent = `Status: ${result.status}`;
        
        if (result.status === 'completed') {
            statusBar.style.background = '#166534'; // green
        } else {
            statusBar.style.background = '#991b1b'; // red
        }

    } catch (err) {
        statusBar.textContent = 'Status: Request Failed';
        statusBar.style.background = '#991b1b';
        stderrEl.textContent = err.message;
    } finally {
        runBtn.disabled = false;
    }
});
