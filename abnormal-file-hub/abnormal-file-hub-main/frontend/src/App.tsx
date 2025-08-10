import React, { useState } from 'react';
import { FileUpload } from './components/FileUpload';
import { FileList } from './components/FileList';
import { useQuery } from '@tanstack/react-query';
import { fileService } from './services/fileService';

function App() {
  const [refreshKey, setRefreshKey] = useState(0);

  const handleUploadSuccess = () => {
    setRefreshKey(prev => prev + 1);
  };

  const { data: stats } = useQuery({
    queryKey: ['stats', refreshKey],
    queryFn: fileService.getStats,
  });

  const formatBytes = (bytes?: number) => {
    if (!bytes) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB'];
    let i = 0;
    let value = bytes;
    while (value >= 1024 && i < units.length - 1) {
      value /= 1024;
      i++;
    }
    return `${value.toFixed(2)} ${units[i]}`;
  };

  return (
    <div className="min-h-screen bg-gray-100">
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
          <h1 className="text-3xl font-bold text-gray-900">Abnormal Security - File Hub</h1>
          <p className="mt-1 text-sm text-gray-500">File management system</p>
        </div>
      </header>
      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          {stats && (
            <div className="grid grid-cols-2 md:grid-cols-3 gap-3 mb-4">
              <div className="bg-white p-3 rounded shadow text-sm">Total files: {stats.total_files}</div>
              <div className="bg-white p-3 rounded shadow text-sm">Duplicates: {stats.duplicates}</div>
              <div className="bg-white p-3 rounded shadow text-sm md:col-span-1 col-span-2">
                Savings: {formatBytes(stats.savings_bytes)}
              </div>
            </div>
          )}
          <div className="space-y-6">
            <div className="bg-white shadow sm:rounded-lg">
              <FileUpload onUploadSuccess={handleUploadSuccess} />
            </div>
            <div className="bg-white shadow sm:rounded-lg">
              <FileList key={refreshKey} />
            </div>
          </div>
        </div>
      </main>
      <footer className="bg-white shadow mt-8">
        <div className="max-w-7xl mx-auto py-4 px-4 sm:px-6 lg:px-8">
          <p className="text-center text-sm text-gray-500">© 2024 File Hub. All rights reserved.</p>
        </div>
      </footer>
    </div>
  );
}

export default App;
