document.addEventListener('DOMContentLoaded', function() {
  // select all <pre><code> elements
  const codeBlocks = document.querySelectorAll('pre > code');

  codeBlocks.forEach(codeBlock => {
    // create container for button
    const copyButtonContainer = document.createElement('div');
    copyButtonContainer.style.position = 'absolute';
    copyButtonContainer.style.top = '0.5em';
    copyButtonContainer.style.right = '0.5em';
    copyButtonContainer.style.opacity = '0'; // initially invisible
    copyButtonContainer.style.transition = 'opacity 0.2s ease-in-out'; // smooth fade-in

    // create copy button
    const copyButton = document.createElement('button');
    copyButton.innerHTML = 'Copy to Clipboard'; // change this to an icon if needed
    copyButton.style.background = 'rgba(255, 255, 255, 0.8)'; // subtle background
    copyButton.style.border = '1px solid #ccc';
    copyButton.style.borderRadius = '4px';
    copyButton.style.padding = '0.25em 0.5em';
    copyButton.style.fontSize = '0.8em';
    copyButton.style.cursor = 'pointer';
    copyButton.style.color = '#333';

    // copy button functionality
    copyButton.addEventListener('click', function() {
      const code = codeBlock.innerText; // set text of code block
      navigator.clipboard.writeText(code)
        .then(() => {
          // optional feedback for the user (e.g., tooltip)
          copyButton.innerHTML = 'content copied';
          setTimeout(() => {
            copyButton.innerHTML = 'Copy to Clipboard';
          }, 2000); // reset after 2 seconds
        })
        .catch(err => {
          console.error('Error copying: ', err);
          copyButton.innerHTML = 'Error!';
        });
    });

    // add button to container
    copyButtonContainer.appendChild(copyButton);

    // find parent <pre> element and add position:relative
    const preBlock = codeBlock.parentNode;
    if (preBlock.tagName === 'PRE') {
      preBlock.style.position = 'relative'; // important for positioning the button
      preBlock.appendChild(copyButtonContainer);

      // event listeners for mouseenter and mouseleave
      preBlock.addEventListener('mouseenter', () => {
        copyButtonContainer.style.opacity = '1';
      });

      preBlock.addEventListener('mouseleave', () => {
        copyButtonContainer.style.opacity = '0';
      });
    } else {
      console.warn('Code block is not directly contained in a <pre> element:', codeBlock);
    }
  });
});
