"""
Setup script for VittoriaDB Python package.
"""

import os
import platform
import urllib.request
import tarfile
import zipfile
from setuptools import setup, find_packages
from setuptools.command.install import install


class PostInstallCommand(install):
    """Custom post-installation to download binaries."""
    
    def run(self):
        install.run(self)
        self.download_binary()
    
    def download_binary(self):
        """Download appropriate binary for the platform."""
        system = platform.system().lower()
        machine = platform.machine().lower()
        
        # Determine binary name
        if system == "darwin":
            arch = "arm64" if machine == "arm64" else "amd64"
            binary_name = f"vittoriadb-darwin-{arch}"
        elif system == "linux":
            arch = "arm64" if machine in ("aarch64", "arm64") else "amd64"
            binary_name = f"vittoriadb-linux-{arch}"
        elif system == "windows":
            binary_name = "vittoriadb-windows-amd64.exe"
        else:
            print(f"Warning: Unsupported platform {system}-{machine}")
            return
        
        # Download URL (GitHub releases)
        version = "v0.1.0"
        url = f"https://github.com/antonellof/VittoriaDB/releases/download/{version}/{binary_name}"
        
        # Local path
        package_dir = os.path.join(self.install_lib, "vittoriadb")
        binary_dir = os.path.join(package_dir, "binaries")
        os.makedirs(binary_dir, exist_ok=True)
        
        binary_path = os.path.join(binary_dir, binary_name)
        
        try:
            print(f"Downloading VittoriaDB binary for {system}-{machine}...")
            urllib.request.urlretrieve(url, binary_path)
            
            # Make executable on Unix systems
            if system in ("darwin", "linux"):
                os.chmod(binary_path, 0o755)
                
            print(f"Downloaded binary to {binary_path}")
            
        except Exception as e:
            print(f"Warning: Failed to download binary: {e}")
            print("You can manually download from https://github.com/antonellof/VittoriaDB/releases")
            print("Or install the Go binary: go install github.com/antonellof/VittoriaDB/cmd/vittoriadb@latest")


# Read README for long description
def read_readme():
    readme_path = os.path.join(os.path.dirname(__file__), "README.md")
    if os.path.exists(readme_path):
        with open(readme_path, "r", encoding="utf-8") as f:
            return f.read()
    return "VittoriaDB - Simple Embedded Vector Database"


# Read version
def read_version():
    version_path = os.path.join(os.path.dirname(__file__), "vittoriadb", "__init__.py")
    with open(version_path, "r", encoding="utf-8") as f:
        for line in f:
            if line.startswith("__version__"):
                return line.split("=")[1].strip().strip('"').strip("'")
    return "0.1.0"


setup(
    name="vittoriadb",
    version=read_version(),
    description="Simple embedded vector database for local AI development with automatic embeddings",
    long_description=read_readme(),
    long_description_content_type="text/markdown",
    author="VittoriaDB Team",
    author_email="team@vittoriadb.dev",
    url="https://github.com/antonellof/VittoriaDB",
    project_urls={
        "Source": "https://github.com/antonellof/VittoriaDB",
        "Tracker": "https://github.com/antonellof/VittoriaDB/issues",
        "Changelog": "https://github.com/antonellof/VittoriaDB/releases",
        "Examples": "https://github.com/antonellof/VittoriaDB/tree/main/examples",
    },
    packages=find_packages(),
    install_requires=[
        "requests>=2.25.0",
    ],
    extras_require={
        "dev": [
            "pytest>=6.0",
            "pytest-cov>=2.0",
            "black>=21.0",
            "mypy>=0.900",
            "flake8>=3.8",
        ],
        "full": [
            "sentence-transformers>=2.0",
            "transformers>=4.0",
            "torch>=1.9",
            "numpy>=1.20",
        ],
        "docs": [
            "sphinx>=4.0",
            "sphinx-rtd-theme>=1.0",
            "myst-parser>=0.15",
        ],
    },
    cmdclass={
        'install': PostInstallCommand,
    },
    entry_points={
        'console_scripts': [
            'vittoriadb-python=vittoriadb.cli:main',
        ],
    },
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "Topic :: Database",
        "Topic :: Scientific/Engineering :: Artificial Intelligence",
        "License :: OSI Approved :: MIT License",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Operating System :: OS Independent",
    ],
    python_requires=">=3.7",
    keywords="vector database, embeddings, similarity search, AI, machine learning, RAG",
    include_package_data=True,
    package_data={
        "vittoriadb": ["binaries/*"],
    },
)
