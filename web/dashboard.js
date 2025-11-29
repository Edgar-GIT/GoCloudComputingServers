document.addEventListener('DOMContentLoaded', function() {
    let currentPath = 'root';
    let selectedItems = new Set();
    let searchQuery = '';
    let allFolders = new Set();
    
    const token = localStorage.getItem('authToken');
    const username = localStorage.getItem('username');
    
    if (!token || !username) {
        window.location.href = 'login.html';
        return;
    }

    document.getElementById('headerUsername').textContent = username;

    function showToast(title, description, variant = 'default') {
        const toastContainer = document.getElementById('toastContainer');
        const toast = document.createElement('div');
        toast.className = `toast ${variant === 'destructive' ? 'destructive' : ''}`;
        
        const titleEl = document.createElement('div');
        titleEl.style.fontWeight = '600';
        titleEl.style.marginBottom = '0.25rem';
        titleEl.textContent = title;
        
        const descEl = document.createElement('div');
        descEl.style.fontSize = '0.875rem';
        descEl.style.color = 'hsl(var(--muted-foreground))';
        descEl.textContent = description;
        
        toast.appendChild(titleEl);
        toast.appendChild(descEl);
        toastContainer.appendChild(toast);

        setTimeout(() => {
            toast.style.animation = 'slide-in 0.3s ease-out reverse';
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    }

    async function apiCall(endpoint, options = {}) {
        const defaultOptions = {
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
                ...options.headers
            }
        };
        
        const response = await fetch(endpoint, { ...defaultOptions, ...options });
        return response;
    }

    function updateSelectionCount() {
        const count = selectedItems.size;
        document.getElementById('downloadCount').textContent = count;
        document.getElementById('deleteCount').textContent = count;
        document.getElementById('downloadBtn').disabled = count === 0;
        document.getElementById('deleteBtn').disabled = count === 0;
    }

    function toggleItemSelection(itemId) {
        if (selectedItems.has(itemId)) {
            selectedItems.delete(itemId);
        } else {
            selectedItems.add(itemId);
        }
        updateSelectionCount();
        updateFileSelection();
    }
    
    function updateFileSelection() {
        const fileItems = document.querySelectorAll('.file-item');
        fileItems.forEach(item => {
            const itemId = item.querySelector('.file-name')?.textContent;
            if (itemId && selectedItems.has(itemId)) {
                item.classList.add('selected');
            } else {
                item.classList.remove('selected');
            }
        });
    }

    function handleItemClick(item) {
        if (item.type === 'folder') {
            currentPath = item.name;
            selectedItems.clear();
            updateSelectionCount();
            loadFiles();
            updateBreadcrumb();
            updateSidebar();
        }
    }

    async function loadFiles() {
        try {
            const pathParam = currentPath === 'root' ? '' : currentPath;
            const response = await apiCall(`/api/files?path=${encodeURIComponent(pathParam)}`);
            
            if (!response.ok) {
                if (response.status === 401) {
                    window.location.href = 'login.html';
                    return;
                }
                throw new Error('Error loading files');
            }
            
            const data = await response.json();
            if (data.success && data.items) {
                renderFilesList(data.items);
            }
        } catch (error) {
            showToast('Error', 'Error loading files', 'destructive');
            console.error(error);
        }
    }

    function updateBreadcrumb() {
        const breadcrumb = document.getElementById('breadcrumb');
        const pathParts = currentPath === 'root' ? ['Home'] : ['Home', currentPath];
        
        breadcrumb.innerHTML = '';
        pathParts.forEach((part, index) => {
            if (index > 0) {
                const separator = document.createElement('svg');
                separator.className = 'breadcrumb-separator icon-chevron-right';
                separator.setAttribute('viewBox', '0 0 24 24');
                separator.innerHTML = '<polyline points="9 18 15 12 9 6"></polyline>';
                breadcrumb.appendChild(separator);
            }
            
            const link = document.createElement('button');
            link.className = 'breadcrumb-link';
            link.textContent = part;
            link.onclick = () => {
                if (index === 0) {
                    currentPath = 'root';
                } else {
                    currentPath = part;
                }
                selectedItems.clear();
                updateSelectionCount();
                loadFiles();
                updateBreadcrumb();
                updateSidebar();
            };
            breadcrumb.appendChild(link);
        });
    }

    function updateSidebar() {
        const homeButton = document.getElementById('homeButton');
        if (homeButton) {
            homeButton.className = `sidebar-button ${currentPath === 'root' ? 'active' : ''}`;
            homeButton.onclick = () => {
                currentPath = 'root';
                selectedItems.clear();
                updateSelectionCount();
                loadFiles();
                updateBreadcrumb();
                updateSidebar();
            };
        }
        
        const sidebarFolders = document.getElementById('sidebarFolders');
        sidebarFolders.innerHTML = '';
        
        Array.from(allFolders).sort().forEach(folderName => {
            const button = document.createElement('button');
            button.className = `sidebar-button ${currentPath === folderName ? 'active' : ''}`;
            button.setAttribute('data-path', folderName);
            button.innerHTML = `
                <img src="gopher-logo.jpg" alt="" class="sidebar-icon">
                ${folderName}
            `;
            button.onclick = () => {
                currentPath = folderName;
                selectedItems.clear();
                updateSelectionCount();
                loadFiles();
                updateBreadcrumb();
                updateSidebar();
            };
            sidebarFolders.appendChild(button);
        });
    }

    function renderFilesList(items) {
        const fileGrid = document.getElementById('fileGrid');
        const emptyState = document.getElementById('emptyState');
        
        items.forEach(item => {
            if (item.type === 'folder') {
                allFolders.add(item.name);
            }
        });
        updateSidebar();
        
        const filteredItems = items.filter(item =>
            item.name.toLowerCase().includes(searchQuery.toLowerCase())
        );

        if (filteredItems.length === 0) {
            fileGrid.style.display = 'none';
            emptyState.style.display = 'block';
            return;
        }

        fileGrid.style.display = 'grid';
        emptyState.style.display = 'none';
        fileGrid.innerHTML = '';

        filteredItems.forEach(item => {
            const fileItem = document.createElement('div');
            const isSelected = selectedItems.has(item.name);
            fileItem.className = `file-item ${isSelected ? 'selected' : ''}`;
            fileItem.onclick = (e) => {
                if (!e.target.closest('.file-item-menu') && !e.target.closest('.dropdown-menu')) {
                    handleItemClick(item);
                }
            };

            fileItem.innerHTML = `
                <div class="file-item-content">
                    <button class="file-item-menu" onclick="event.stopPropagation(); showDropdown(event, '${item.id}', '${item.name}', '${item.type}')">
                        <svg class="btn-icon icon-more-vertical" viewBox="0 0 24 24">
                            <circle cx="12" cy="12" r="1"></circle>
                            <circle cx="12" cy="5" r="1"></circle>
                            <circle cx="12" cy="19" r="1"></circle>
                        </svg>
                    </button>
                    ${item.type === 'folder' 
                        ? `<img src="gopher-logo.jpg" alt="Folder" class="file-icon">`
                        : `<svg class="file-icon-svg icon-file" viewBox="0 0 24 24">
                            <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"></path>
                            <polyline points="13 2 13 9 20 9"></polyline>
                        </svg>`
                    }
                    <div class="text-center w-100">
                        <p class="file-name">${item.name}</p>
                        <p class="file-meta">${item.size || item.modified}</p>
                    </div>
                </div>
            `;

            fileGrid.appendChild(fileItem);
        });
    }

    let currentDropdownItem = null;
    
    window.showDropdown = function(event, itemId, itemName, itemType) {
        event.stopPropagation();
        const dropdown = document.getElementById('dropdownMenu');
        
        currentDropdownItem = { id: itemId, name: itemName, type: itemType };
        const isSelected = selectedItems.has(itemName);
        
        document.getElementById('dropdownSelect').textContent = isSelected ? 'Deselect' : 'Select';
        
        dropdown.style.display = 'block';
        dropdown.style.left = event.pageX + 'px';
        dropdown.style.top = event.pageY + 'px';
        
        setTimeout(() => {
            document.addEventListener('click', closeDropdown);
        }, 0);
    };

    function closeDropdown() {
        document.getElementById('dropdownMenu').style.display = 'none';
        document.removeEventListener('click', closeDropdown);
        currentDropdownItem = null;
    }

    document.getElementById('dropdownSelect').onclick = function(e) {
        e.stopPropagation();
        if (currentDropdownItem) {
            toggleItemSelection(currentDropdownItem.name);
        }
        closeDropdown();
    };

    document.getElementById('dropdownRename').onclick = async function(e) {
        e.stopPropagation();
        if (!currentDropdownItem) {
            closeDropdown();
            return;
        }
        
        const newName = prompt('New name:', currentDropdownItem.name);
        if (!newName || newName === currentDropdownItem.name) {
            closeDropdown();
            return;
        }
        
        try {
            const response = await apiCall('/api/files/rename', {
                method: 'POST',
                body: JSON.stringify({
                    path: currentPath === 'root' ? '' : currentPath,
                    oldName: currentDropdownItem.name,
                    newName: newName
                })
            });
            
            if (response.ok) {
                showToast('Renamed', `Renamed to ${newName}`);
                loadFiles();
            } else {
                const data = await response.json();
                showToast('Error', data.error || 'Error renaming', 'destructive');
            }
        } catch (error) {
            showToast('Error', 'Error renaming', 'destructive');
        }
        closeDropdown();
    };

    document.getElementById('dropdownDelete').onclick = function(e) {
        e.stopPropagation();
        if (currentDropdownItem) {
            handleDelete([currentDropdownItem.name]);
        }
        closeDropdown();
    };

    document.getElementById('createFolderBtn').onclick = async function() {
        const folderName = prompt('Folder name:');
        if (!folderName) return;
        
        try {
            const response = await apiCall('/api/files/folder', {
                method: 'POST',
                body: JSON.stringify({
                    path: currentPath === 'root' ? '' : currentPath,
                    folderName: folderName
                })
            });
            
            if (response.ok) {
                showToast('Folder Created', `Folder "${folderName}" created`);
                loadFiles();
            } else {
                const data = await response.json();
                showToast('Error', data.error || 'Error creating folder', 'destructive');
            }
        } catch (error) {
            showToast('Error', 'Error creating folder', 'destructive');
        }
    };

    document.getElementById('uploadBtn').onclick = function() {
        const input = document.createElement('input');
        input.type = 'file';
        input.multiple = true;
        input.onchange = async function(e) {
            const files = e.target.files;
            if (files.length === 0) return;
            
            const formData = new FormData();
            for (let i = 0; i < files.length; i++) {
                formData.append('files', files[i]);
            }
            
            try {
                const pathParam = currentPath === 'root' ? '' : currentPath;
                const response = await fetch(`/api/files/upload?path=${encodeURIComponent(pathParam)}`, {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${token}`
                    },
                    body: formData
                });
                
                const data = await response.json();
                if (response.ok && data.success) {
                    showToast('Upload Successful', `${data.uploaded} file(s) uploaded`);
                    loadFiles();
                } else {
                    showToast('Error', data.error || 'Error uploading files', 'destructive');
                }
            } catch (error) {
                showToast('Error', 'Error uploading files', 'destructive');
            }
        };
        input.click();
    };

    document.getElementById('downloadBtn').onclick = async function() {
        if (selectedItems.size === 0) {
            showToast('No Selection', 'Please select files to download', 'destructive');
            return;
        }
        
        for (const fileName of selectedItems) {
            const pathParam = currentPath === 'root' ? '' : currentPath;
            const url = `/api/files/download?path=${encodeURIComponent(pathParam)}&name=${encodeURIComponent(fileName)}&token=${encodeURIComponent(token)}`;
            window.open(url, '_blank');
        }
        
        showToast('Download', `Downloading ${selectedItems.size} item(s)`);
    };

    async function handleDelete(itemNames) {
        if (itemNames.length === 0) {
            showToast('No Selection', 'Please select files to delete', 'destructive');
            return;
        }
        
        if (!confirm(`Are you sure you want to delete ${itemNames.length} item(s)?`)) {
            return;
        }
        
        try {
            const response = await apiCall('/api/files', {
                method: 'DELETE',
                body: JSON.stringify({
                    path: currentPath === 'root' ? '' : currentPath,
                    names: itemNames
                })
            });
            
            if (response.ok) {
                showToast('Deleted', `Deleted ${itemNames.length} item(s)`);
                itemNames.forEach(name => selectedItems.delete(name));
                updateSelectionCount();
                loadFiles();
            } else {
                const data = await response.json();
                showToast('Error', data.error || 'Error deleting', 'destructive');
            }
        } catch (error) {
            showToast('Error', 'Error deleting', 'destructive');
        }
    }

    document.getElementById('deleteBtn').onclick = function() {
        handleDelete(Array.from(selectedItems));
    };

    document.getElementById('logoutBtn').onclick = async function() {
        try {
            await apiCall('/api/logout', { method: 'POST' });
        } catch (error) {
            console.error('Logout error:', error);
        }
        
        localStorage.removeItem('authToken');
        localStorage.removeItem('username');
        showToast('Logged Out', 'You have been successfully logged out');
        setTimeout(() => {
            window.location.href = 'login.html';
        }, 500);
    };

    document.getElementById('searchInput').addEventListener('input', function(e) {
        searchQuery = e.target.value;
        loadFiles();
    });

    updateBreadcrumb();
    loadFiles();
    updateSelectionCount();
});
