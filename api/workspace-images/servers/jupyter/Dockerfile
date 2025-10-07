FROM kloudlite/workspace-base:latest

USER root

# Install Python and pip
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

USER workspace

# Install JupyterLab
RUN pip3 install --user jupyterlab notebook ipykernel && \
    python3 -m ipykernel install --user --name python3

# Configure Jupyter
RUN mkdir -p /home/workspace/.jupyter && \
    echo "c.ServerApp.token = ''" > /home/workspace/.jupyter/jupyter_server_config.py && \
    echo "c.ServerApp.password = ''" >> /home/workspace/.jupyter/jupyter_server_config.py && \
    echo "c.ServerApp.allow_origin = '*'" >> /home/workspace/.jupyter/jupyter_server_config.py && \
    echo "c.ServerApp.disable_check_xsrf = True" >> /home/workspace/.jupyter/jupyter_server_config.py

EXPOSE 8888

CMD ["/home/workspace/.local/bin/jupyter", "lab", "--ip=0.0.0.0", "--port=8888", "--no-browser", "--notebook-dir=/workspace"]
