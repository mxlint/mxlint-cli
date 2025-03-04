(function() {
    const passwordInput = document.getElementById("passwordInput");
    const passwordToggle = document.getElementById("passwordToggle");

    function togglePassword() {
        if (passwordToggle.classList.contains("glyphicon-eye-close"))
            passwordToggle.classList.replace("glyphicon-eye-close", "glyphicon-eye-open");
        else passwordToggle.classList.replace("glyphicon-eye-open", "glyphicon-eye-close");

        if (passwordInput.type === "password") passwordInput.type = "text";
        else passwordInput.type = "password";
    }

    passwordToggle.addEventListener("click", togglePassword);
})();