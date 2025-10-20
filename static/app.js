function refreshStrings() {
    var printableLength = document.getElementById("p").value;
    if (printableLength > 99) {
        printableLength = 99;
        document.getElementById("p").value = 99;
    }
    if (printableLength < 1) {
        printableLength = 1;
        document.getElementById("p").value = 1;
    }
    var alphanumericLength = document.getElementById("a").value;
    if (alphanumericLength > 99) {
        alphanumericLength = 99;
        document.getElementById("a").value = 99;
    }
    if (alphanumericLength < 1) {
        alphanumericLength = 1;
        document.getElementById("a").value = 1;
    }
    var url = "/json?p=" + printableLength + "&a=" + alphanumericLength;

    fetch(url, { cache: 'no-store' })
        .then(response => response.json())
        .then(data => {
            document.getElementById("printable-string").textContent = data.printable.string;
            document.getElementById("alphanumeric-string").textContent = data.alphanumeric.string;
        });
}

// Reload the page without using the browser cache
function refreshNoCache(event) {
    if (event) event.preventDefault();
    // Use location.replace so the back button isn't polluted with cache-busting URLs.
    // If the current path ends with index.html, strip it so we reload the directory root.
    var path = window.location.pathname.replace(/index\.html$/, '/');
    if (path === '') path = '/';
    var url = path + '?_=' + Date.now();
    window.location.replace(url);
}

function copyToClipboard(elementId, buttonId) {
    var text = document.getElementById(elementId).textContent;
    var button = document.getElementById(buttonId);

    // Try modern clipboard API first
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(function () {
            showCopySuccess(button);
        }).catch(function (err) {
            console.error('Clipboard API failed:', err);
            fallbackCopy(text, button);
        });
    } else {
        // Fallback for older browsers or HTTP environments
        fallbackCopy(text, button);
    }
}

function showCopySuccess(button) {
    var originalText = button.textContent;
    button.textContent = '✓ Copied';
    button.classList.add('copied');

    setTimeout(function () {
        button.textContent = originalText;
        button.classList.remove('copied');
    }, 2000);
}

function fallbackCopy(text, button) {
    // Check if execCommand is supported
    if (!document.queryCommandSupported || !document.queryCommandSupported('copy')) {
        button.textContent = '✗ Failed';
        setTimeout(function () {
            button.textContent = 'Copy';
        }, 2000);
        return;
    }

    var textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.position = 'fixed';
    textarea.style.left = '-9999px';
    textarea.style.top = '-9999px';
    textarea.style.opacity = '0';
    textarea.setAttribute('aria-hidden', 'true');
    textarea.setAttribute('readonly', '');

    var added = false;
    try {
        document.body.appendChild(textarea);
        added = true;
        textarea.select();

        var successful = document.execCommand('copy');
        if (successful) {
            showCopySuccess(button);
        } else {
            button.textContent = '✗ Failed';
            setTimeout(function () {
                button.textContent = 'Copy';
            }, 2000);
        }
    } catch (err) {
        console.error('Fallback copy failed:', err);
        button.textContent = '✗ Failed';
        setTimeout(function () {
            button.textContent = 'Copy';
        }, 2000);
    } finally {
        if (added && textarea.parentNode) {
            document.body.removeChild(textarea);
        }
    }
}
