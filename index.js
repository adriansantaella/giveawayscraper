document.addEventListener('DOMContentLoaded', (e) => {
    const numberOfPagesInput = document.getElementById('numberOfPages');
	const fetchResultsBtn = document.getElementById('fetchResultsBtn');
    
    numberOfPagesInput.addEventListener('click', (e) => {
    });
    
    fetchResultsBtn.addEventListener('click', (e) => {
        let val = numberOfPagesInput.value;
        handleSubmit(val);
    });

    let scrapedItems = [];

	let handleSubmit = async (numOfPages) => {
		try {
			const response = await fetch(`http://localhost:8080/scrape-data?numpages=${numOfPages}`);
			if (!response.ok) {
				throw new Error('Response came back with errors. Try again later...');
			}
			const data = await response.json();
			scrapedItems = data.items;
		} catch (error) {
			console.error('There was a problem with fetching the data...');
		}
	};
});