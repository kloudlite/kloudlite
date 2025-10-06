FROM kloudlite/workspace-base:latest

USER root

# Install Python and pip via Nix
RUN . /home/workspace/.nix-profile/etc/profile.d/nix.sh && \
    nix-env -iA nixpkgs.python311 nixpkgs.python311Packages.pip

USER workspace

# Install JupyterLab
RUN . /home/workspace/.nix-profile/etc/profile.d/nix.sh && \
    pip install --user jupyterlab notebook ipykernel && \
    python -m ipykernel install --user --name python3

# Configure Jupyter
RUN mkdir -p /home/workspace/.jupyter && \
    echo "c.ServerApp.token = ''" > /home/workspace/.jupyter/jupyter_server_config.py && \
    echo "c.ServerApp.password = ''" >> /home/workspace/.jupyter/jupyter_server_config.py && \
    echo "c.ServerApp.allow_origin = '*'" >> /home/workspace/.jupyter/jupyter_server_config.py && \
    echo "c.ServerApp.disable_check_xsrf = True" >> /home/workspace/.jupyter/jupyter_server_config.py

EXPOSE 8888

CMD ["/home/workspace/.local/bin/jupyter", "lab", "--ip=0.0.0.0", "--port=8888", "--no-browser", "--notebook-dir=/workspace"]
