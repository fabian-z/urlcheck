// JavaScript code for changing the language of the web page
// Get the elements with lang attribute
var elements = document.querySelectorAll("[lang]");

// Get the buttons for switching the language
var buttons = document.querySelectorAll(".language-switcher button");

// Define a function to get the browser language
function getBrowserLanguage() {
    // Get the language from the navigator object
    var language = navigator.language || navigator.userLanguage;
    // Use only the first two characters of the language code
    language = language.substring(0, 2);
    
    // If the language is not English, German or Chinese, default to English
    if (language != "en" && language != "de" && language != "zh") {
        language = "en";
    }

    // Return the language
    return language;
}

// Define a function to change the language of the web page
function changeLanguage(lang) {
    // Loop through the elements with lang attribute
    elements.forEach(function(element) {
				
        // If the element lang matches the parameter, show the element
       
        if (element.lang == lang) {
            console.log(element, lang);
		    if (element.tagName == "TITLE") {
				console.log("Changing title to " + element.innerHTML);
			    document.title = element.innerHTML;
		    } else {
			    element.hidden = false;
		    }
        }
        // Otherwise, hide the element
        else {
            element.hidden = true;
        }      
        
    });

    // Loop through the buttons
    buttons.forEach(function(button) {
        // If the button id matches the parameter, add the active class
        if (button.id == lang) {
            button.classList.add("active");
        }
        // Otherwise, remove the active class
        else {
            button.classList.remove("active");
        }
    });
}

// Add click event listeners to the buttons
buttons.forEach(function(button) {
    button.addEventListener("click", function() {
        // Get the language code from the button id
        var lang = button.id;

        // Call the changeLanguage function with the language code
        changeLanguage(lang);
    });
});

// Call the changeLanguage function with the browser language when the page loads
window.addEventListener("load", function() {
    // Get the browser language
    var lang = getBrowserLanguage();

    // Call the changeLanguage function with the browser language
    changeLanguage(lang);
});
