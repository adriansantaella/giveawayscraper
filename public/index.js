document.addEventListener("DOMContentLoaded", (e) => {
    const numberOfPagesInput = document.getElementById("numberOfPages");
    const fetchResultsBtn = document.getElementById("fetchResultsBtn");
    const giveawayGrid = document.getElementById("giveawayGrid");

    
    numberOfPagesInput.addEventListener("click", (e) => {});
    
    fetchResultsBtn.addEventListener("click", (e) => {
        const spinner = document.createElement("div");
        giveawayGrid.innerHTML = "";

        spinner.classList.add("no-entries");
        spinner.innerHTML = '<i class="fa-solid fa-spinner fa-pulse"></i>';
        
        giveawayGrid.appendChild(spinner);

        let val = numberOfPagesInput.value;
        handleSubmit(val);
    });

    let handleSubmit = async (numOfPages) => {
        try {
            const response = await fetch(`/api/scrape?numpages=${numOfPages}`);

            if (!response.ok) {
                throw new Error(
                    "Response came back with errors. Try again later..."
                );
            }
            
            await response.json().then((data) => {
                if(data) {
                    fillGiveawayGrid(data.items);
                } else {
                    showEmptyResults();
                }
            });

        } catch (error) {
            console.error(
                "There was a problem with fetching the data...",
                error
            );
        }
    };

    let fillGiveawayGrid = (data) => {
        giveawayGrid.innerHTML = "";

        data.forEach((item) => {
            let newEl = document.createElement("div");
            newEl.classList.add("giveawayItem");

            newEl.innerHTML = `<img src="${item.ImageURL}" alt="${item.Name}" />
                    <div class="giveawayDesc">
                        <a href="${item.URL}" target="_blank">${item.Name}</a>
                        <span>expires on ${item.ExpirationDate}</span>
                    </div>`;

            giveawayGrid.appendChild(newEl);
        });
    };

    let showEmptyResults = () => {
        giveawayGrid.innerHTML = "";
        const noEntries = document.createElement("div");
        noEntries.classList.add("no-entries");
        noEntries.innerText = "No entries found.";

        giveawayGrid.appendChild(noEntries);
    }
});
